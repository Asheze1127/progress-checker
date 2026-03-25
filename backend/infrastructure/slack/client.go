package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const slackPostMessageURL = "https://slack.com/api/chat.postMessage"

// Client implements the SlackClient interface using the Slack Web API.
type Client struct {
	botToken   string
	httpClient *http.Client
}

// NewClient creates a new Slack API client.
func NewClient(botToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		botToken:   botToken,
		httpClient: httpClient,
	}
}

// postMessageRequest represents the request body for chat.postMessage.
type postMessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// postMessageResponse represents the response from chat.postMessage.
type postMessageResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// PostMessage sends a message to a Slack channel using chat.postMessage.
func (c *Client) PostMessage(ctx context.Context, channelID string, text string) error {
	reqBody := postMessageRequest{
		Channel: channelID,
		Text:    text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackPostMessageURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d: %s", resp.StatusCode, string(body))
	}

	var slackResp postMessageResponse
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return fmt.Errorf("failed to parse slack response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return nil
}
