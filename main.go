package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/poornima-krishnasamy/cloud-platform-slack-metrics/pkg/metrics"
)

func main() {

	var (
		botToken    = flag.String("token", os.Getenv("SLACK_BOT_OAUTH_TOKEN"), "Bot token from Slack App: Answerbot.")
		channelName = flag.String("channel", os.Getenv("SLACK_CHANNEL"), "Name of Slack channel to get the response metrics.")
		teamName    = flag.String("cpteam", os.Getenv("SLACK_CP_TEAM"), "Name of team to exclude from when calculating response time.")
	)

	flag.Parse()

	s := metrics.NewSlack(*botToken)

	err := s.TeamMembers(*teamName)
	if err != nil {
		fmt.Println(err)
		return
	}

	channelID, err := s.GetChannelID(*channelName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("ChannelID %s\n", channelID)

}
