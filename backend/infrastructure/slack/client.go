package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	slackposter "github.com/Asheze1127/progress-checker/backend/application/service/slack_poster"
	threadfetcher "github.com/Asheze1127/progress-checker/backend/application/service/thread_fetcher"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// Compile-time checks.
var (
	_ slackposter.SlackClient          = (*Client)(nil)
	_ threadfetcher.SlackThreadFetcher = (*Client)(nil)
)

// Client implements the SlackClient and SlackThreadFetcher interfaces
// using the official slack-go library.
type Client struct {
	api *slack.Client
}

// NewClient creates a new Slack API client.
func NewClient(botToken string) *Client {
	return &Client{
		api: slack.New(botToken),
	}
}

// API returns the underlying slack-go client for use by other infrastructure types.
func (c *Client) API() *slack.Client {
	return c.api
}

// PostMessage sends a message to a Slack channel using chat.postMessage.
func (c *Client) PostMessage(ctx context.Context, channelID string, text string) error {
	_, _, err := c.api.PostMessageContext(ctx, channelID, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("failed to post slack message: %w", err)
	}
	return nil
}

// SlackUserInfo holds basic user info from the Slack API.
type SlackUserInfo struct {
	ID       string
	RealName string
	Email    string
}

// GetUserInfo fetches user information from the Slack API.
func (c *Client) GetUserInfo(ctx context.Context, userID string) (*SlackUserInfo, error) {
	user, err := c.api.GetUserInfoContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get slack user info: %w", err)
	}
	return &SlackUserInfo{
		ID:       user.ID,
		RealName: user.RealName,
		Email:    user.Profile.Email,
	}, nil
}

// FetchThreadMessages retrieves all replies in a Slack thread.
func (c *Client) FetchThreadMessages(ctx context.Context, channelID, threadTS string) ([]pkgslack.ThreadMessage, error) {
	msgs, _, _, err := c.api.GetConversationRepliesContext(ctx, &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	result := make([]pkgslack.ThreadMessage, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, pkgslack.ThreadMessage{
			User: m.User,
			Text: m.Text,
			TS:   m.Timestamp,
		})
	}
	return result, nil
}
