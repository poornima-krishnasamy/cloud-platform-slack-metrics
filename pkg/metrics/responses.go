package metrics

import (
	"fmt"
	"log"

	"github.com/poornima-krishnasamy/cloud-platform-slack-metrics/pkg/utils"
	"github.com/slack-go/slack"
)

type CPSlack struct {
	client    *slack.Client
	cpMembers []string
}

func NewSlack(token string) *CPSlack {
	cpslack := new(CPSlack)
	cpslack.client = slack.New(token)
	return cpslack
}

// Get the list of users of cloud-platform-team group.
// This will be used to find the first reply from the cloud platform team

func (cpslack *CPSlack) TeamMembers(teamName string) error {

	var cpMemberslist []string

	userGroups, err := cpslack.client.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
	if err != nil {
		fmt.Printf("Unexpected error: %s", err)
		return err
	}

	if len(userGroups) < 1 {
		fmt.Printf("No usergroups available")
		return err
	}

	for _, group := range userGroups {
		if group.Handle == teamName {
			cpMemberslist = group.Users
			break
		}

	}
	cpslack.cpMembers = cpMemberslist
	return nil
}

// Get the list of channels and get the channel ID for the expected channelName
func (cpslack *CPSlack) GetChannelID(channelName string) (string, error) {

	var allChannels []slack.Channel
	var channelID string

	// Initial request
	initChans, initCur, err := cpslack.client.GetConversations(
		&slack.GetConversationsParameters{
			ExcludeArchived: true,
			Limit:           1000,
			Types: []string{
				"public_channel",
			},
		},
	)
	if err != nil {
		log.Fatalln("Unexpected Error:", err)
		return "", err
	}

	allChannels = append(allChannels, initChans...)

	// Paginate over additional channels
	nextCur := initCur
	for nextCur != "" {
		channels, cursor, err := cpslack.client.GetConversations(
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
			log.Fatalln("Unexpected Error:", err)
			return "", err
		}

		allChannels = append(allChannels, channels...)
		nextCur = cursor
	}

	for _, channel := range allChannels {
		if channel.Name == channelName {
			channelID = channel.ID
			break
		}
	}
	return channelID, nil
}

// 	// Get the list of threads for a channel
// 	// There is a 50+ rate limit for conversations.history. If the limit has to be increased add a delay
// 	// https://api.slack.com/docs/rate-limits#tier_t3

func (cpslack *CPSlack) GetMessagesHistory(channelID string) ([]slack.Message, error) {

	// TODO Accomodate pagination
	history, err := cpslack.client.GetConversationHistory(
		&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     100,
			Inclusive: false,
		},
	)
	if err != nil {
		log.Fatalln("Unexpected Error %w", err)
		return nil, err
	}
	return history.Messages, nil

}

// Get the message Timestamp for a given slack message

func (cpslack *CPSlack) GetValidMessageTimestamp(message slack.Message, channelID string) (string, error) {

	threadTime, _ := utils.TimeSectoTime(message.Timestamp)
	msgTimeLocal := utils.LocalTime(threadTime)
	//Ignore threads which are posted by a member of cloud platform team. This could be updates to users
	if utils.Contains(cpslack.cpMembers, message.User) {
		//	log.Printf("Thread posted by member of CP team. \nThread : %v\n", message)
		return "", nil
	} else if message.ReplyCount == 0 {
		// Ignore threads which has no reply
		//log.Printf("Thread has no response. Possibly a PR review\nThread: %v\n", message)
		return "", nil
	} else if msgTimeLocal.Hour() < 10 || msgTimeLocal.Hour() > 16 {
		// Ignore threads posted out of hours anything before 10am and after 5pm
		//	log.Printf("Thread posted out of hours.\nThread: %v\n", message)
		return "", nil
	} else if message.ThreadTimestamp != "" && message.ThreadTimestamp == message.Timestamp {
		// When the message timestamp and thread timestamp are the same, we
		// have a parent message. This means it contains a thread with replies.
		//log.Printf("Thread: %v\n", message)
		return message.Timestamp, nil
	} else {
		//log.Printf("Thread not in any conditions.\nThread: %v\n", message)
		return "", nil
	}
}

// Given a threadTimestamp, get all replies for that thread and
// return the first cp member response Timestamp
func (cpslack *CPSlack) GetFirstReplyTimestamp(channelID string, threadTimestamp string) (string, error) {
	// limited  to first 10 responses assuming it will have the first cp member response
	replies, _, _, err := cpslack.client.GetConversationReplies(
		&slack.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: threadTimestamp,
			Limit:     100,
		},
	)
	if err != nil {
		log.Fatalln("Unexpected Error %w", err)
		return "", nil
	}

	for _, reply := range replies {
		// Because the conversations api returns an entire thread (a
		// message plus all the messages in reply), we need to check if
		// one of the replies isn't the parent that we started with.
		if reply.ThreadTimestamp != "" && reply.ThreadTimestamp == reply.Timestamp {
			//log.Printf("Thread reply is the  parent.\n")
			continue
		} else if !utils.Contains(cpslack.cpMembers, reply.User) {
			//log.Printf("Thread not in any conditions.\n")
			continue
		} else {
			//Get the first response and break
			return reply.Timestamp, nil
		}

	}
	return "", nil
}
