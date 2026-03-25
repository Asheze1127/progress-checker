package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Client implements the SlackClient interface using the official slack-go library.
type Client struct {
	api *slack.Client
}

// NewClient creates a new Slack API client.
func NewClient(botToken string) *Client {
	return &Client{
		api: slack.New(botToken),
	}
}

// PostMessage sends a message to a Slack channel using chat.postMessage.
func (c *Client) PostMessage(_ context.Context, channelID string, text string) error {
	_, _, err := c.api.PostMessage(channelID, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("failed to post slack message: %w", err)
	}
	return nil
}
