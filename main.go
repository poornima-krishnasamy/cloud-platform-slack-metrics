package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func main() {

	var (
		oAuthToken    = flag.String("token", os.Getenv("SLACK_BOT_OAUTH_TOKEN"), "Bot token from Slack App: Answerbot.")
		channelName   = flag.String("channel", os.Getenv("SLACK_CHANNEL"), "Name of Slack channel to get the response metrics.")
		channelID     string
		cpMemberslist []string
		responseTimes []int
	)

	api := slack.New(*oAuthToken, slack.OptionDebug(true))

	// Get the list of channels and get the channel ID for the expected channelName
	var allChannels []slack.Channel
	initChans, initCur, err := api.GetConversations(
		&slack.GetConversationsParameters{
			ExcludeArchived: true,
			Limit:           1000,
			Types: []string{
				"public_channel",
			},
		},
	)
	if err != nil {

		return
	}

	allChannels = append(allChannels, initChans...)

	// Paginate over additional channels
	nextCur := initCur
	for nextCur != "" {
		channels, cursor, err := api.GetConversations(
			&slack.GetConversationsParameters{
				Cursor:          nextCur,
				ExcludeArchived: true,
				Limit:           1000,
				Types: []string{
					"public_channel",
				},
			},
		)
		if err != nil {
			log.Fatalf("Unexpected error: %s", err)
		}

		allChannels = append(allChannels, channels...)
		nextCur = cursor
	}

	for _, channel := range allChannels {
		if channel.Name == *channelName {
			channelID = channel.ID
			break
		}
	}

	// Get the list of users of cloud-platform-team group.
	// This will be used to find the first reply from the cloud platform team

	userGroups, err := api.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
	if err != nil {
		fmt.Printf("Unexpected error: %s", err)
		return
	}

	if len(userGroups) < 1 {
		fmt.Printf("No usergroups available")
		return
	}

	for _, group := range userGroups {
		if group.Handle == "cloud-platform-team" {
			cpMemberslist = group.Users
			break
		}

	}

	// Get the list of threads for a channel
	// Add pagination by checking hasmore and sending cursor TODO
	// There is a 50+ rate limit for conversations.history. If the limit has to be increased add a delay
	// https://api.slack.com/docs/rate-limits#tier_t3

	history, err := api.GetConversationHistory(
		&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     100,
			Inclusive: false,
		},
	)
	if err != nil {
		fmt.Println("Unexpected Error %w", err)
	}

	i := 0
	for _, message := range history.Messages {
		i++
		messageTimestamp, _ := strconv.Atoi(message.Timestamp[0:strings.LastIndex(message.Timestamp, ".")])
		msgTimeLocal := time.Unix(int64(messageTimestamp), 0)
		fmt.Printf("Msg %v - Message Time%v\n", i, msgTimeLocal)

		//Ignore threads which are posted by a member of cloud platform team. This could be updates to users
		if contains(cpMemberslist, message.User) {
			fmt.Println("Thread posted by member of CP team")
			continue
		}
		// Ignore threads which has no reply
		if message.ReplyCount == 0 {
			fmt.Println("Thread has no response. Possibly a PR review")
			continue
		}
		// Ignore threads posted out of hours anything before 10am and after 5pm

		if msgTimeLocal.Hour() < 10 || msgTimeLocal.Hour() > 16 {
			fmt.Println("Thread posted out of hours")
			continue
		}

		// When the message timestamp and thread timestamp are the same, we
		// have a parent message. This means it contains a thread with replies.

		if message.ThreadTimestamp != "" && message.ThreadTimestamp == message.Timestamp {
			// Get all replies from the thread. The api returns the replies with the latest
			// as the first element.
			// Add pagination by checking hasmore and sending cursor TODO
			replies, _, _, err := api.GetConversationReplies(
				&slack.GetConversationRepliesParameters{
					ChannelID: channelID,
					Timestamp: message.ThreadTimestamp,
					Limit:     100,
				},
			)
			if err != nil {
				fmt.Println("Unexpected Error %w", err)
				return
			}

			for _, reply := range replies {
				// Because the conversations api returns an entire thread (a
				// message plus all the messages in reply), we need to check if
				// one of the replies isn't the parent that we started with.
				if reply.ThreadTimestamp != "" && reply.ThreadTimestamp == reply.Timestamp {
					continue
				} else if !contains(cpMemberslist, reply.User) {
					continue
				} else {
					//Get the first response and break

					replyTimestamp, _ := strconv.Atoi(reply.Timestamp[0:strings.LastIndex(reply.Timestamp, ".")])
					responseTimes = append(responseTimes, replyTimestamp-messageTimestamp)
					// The below are for just output prints.
					replyTimeLocal := time.Unix(int64(replyTimestamp), 0)
					fmt.Printf("Message Time,%v,Reply Time,%v\n", msgTimeLocal, replyTimeLocal)
					diff := replyTimeLocal.Sub(msgTimeLocal)
					fmt.Printf("Response time is %v\n", diff)

					break
				}

			}
		}

	}
	totalTime := 0
	for _, time := range responseTimes {
		totalTime = totalTime + time
	}
	average := totalTime / len(responseTimes)
	fmt.Printf("Average response time is %v\n", time.Duration(average)*time.Second)

}
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
