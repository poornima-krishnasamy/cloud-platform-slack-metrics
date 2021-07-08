package metrics

import (
	"fmt"

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
	fmt.Println(s.cpMembers)
	return nil
}
