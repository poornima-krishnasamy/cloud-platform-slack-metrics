package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/poornima-krishnasamy/cloud-platform-slack-metrics/pkg/metrics"
	"github.com/poornima-krishnasamy/cloud-platform-slack-metrics/pkg/utils"
)

func main() {

	var (
		botToken    = flag.String("token", os.Getenv("SLACK_BOT_OAUTH_TOKEN"), "Bot token from Slack App: Answerbot.")
		channelName = flag.String("channel", os.Getenv("SLACK_CHANNEL"), "Name of Slack channel to get the response metrics.")
		teamName    = flag.String("cpteam", os.Getenv("SLACK_CP_TEAM"), "Name of team to exclude from when calculating response time.")
		// TODO: Use start and end time to get response for particular duration
		responseTimes []int
	)

	flag.Parse()

	cpslack := metrics.NewSlack(*botToken)

	err := cpslack.TeamMembers(*teamName)
	if err != nil {
		fmt.Println(err)
		return
	}

	channelID, err := cpslack.GetChannelID(*channelName)
	if err != nil || channelID == "" {
		fmt.Println(err)
		return
	}

	messages, err := cpslack.GetMessagesHistory(channelID)
	if err != nil || messages == nil {
		fmt.Println(err)
		return
	}

	for _, message := range messages {
		threadTimestamp, err := cpslack.GetValidMessageTimestamp(message, channelID)
		if err != nil {
			fmt.Println(err)
			return
		}
		// If there is a timestamp, then it is
		if threadTimestamp != "" {
			replyTimestamp, err := cpslack.GetFirstReplyTimestamp(channelID, threadTimestamp)
			if err != nil {
				fmt.Println(err)
				return
			}
			if replyTimestamp != "" {
				threadTime, _ := utils.TimeSectoTime(threadTimestamp)
				replyTime, _ := utils.TimeSectoTime(replyTimestamp)
				responseTimes = append(responseTimes, replyTime-threadTime)

				// The below are for just output prints.
				replyTimeLocal := utils.LocalTime(replyTime)
				threadTimeLocal := utils.LocalTime(threadTime)
				fmt.Printf("Message Time,%v,Reply Time,%v\n", threadTimeLocal, replyTimeLocal)
				diff := replyTimeLocal.Sub(threadTimeLocal)
				fmt.Printf("Response time is %v\n", diff)
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
