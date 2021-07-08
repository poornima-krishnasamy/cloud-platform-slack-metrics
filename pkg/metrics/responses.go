package metrics

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
)

type Slack struct {
	client    *slack.Client
	cpMembers []string
}

func NewSlack(token string) *Slack {
	s := new(Slack)
	s.client = slack.New(token)
	return s
}

// Get the list of users of cloud-platform-team group.
// This will be used to find the first reply from the cloud platform team

func (s *Slack) TeamMembers(teamName string) error {

	var cpMemberslist []string

	userGroups, err := s.client.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
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
	s.cpMembers = cpMemberslist
	return nil
}

// Get the list of channels and get the channel ID for the expected channelName
func (s *Slack) GetChannelID(channelName string) (string, error) {

	var allChannels []slack.Channel
	var channelID string

	// Initial request
	initChans, initCur, err := s.client.GetConversations(
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
		channels, cursor, err := s.client.GetConversations(
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
