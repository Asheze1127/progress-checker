package slack

// SlackEvent represents the top-level envelope for all Slack Events API payloads.
type SlackEvent struct {
	Token     string `json:"token"`
	Type      string `json:"type"`
	Challenge string `json:"challenge,omitempty"`

	// Fields present when Type == "event_callback"
	TeamID    string         `json:"team_id,omitempty"`
	EventID   string         `json:"event_id,omitempty"`
	EventTime int64          `json:"event_time,omitempty"`
	Event     EventCallback  `json:"event,omitempty"`
	Actions   []ActionDetail `json:"actions,omitempty"`

	// Fields present for message_action (shortcut) payloads
	CallbackID string  `json:"callback_id,omitempty"`
	TriggerID  string  `json:"trigger_id,omitempty"`
	User       IDField `json:"user,omitempty"`
	Channel    IDField `json:"channel,omitempty"`
	MessageTS  string  `json:"message_ts,omitempty"`
	Message    Message `json:"message,omitempty"`
}

// EventCallback represents the inner event object within an event_callback envelope.
type EventCallback struct {
	Type string `json:"type"`

	// Fields for reaction_added events
	User     string       `json:"user,omitempty"`
	Reaction string       `json:"reaction,omitempty"`
	Item     ReactionItem `json:"item,omitempty"`
	EventTS  string       `json:"event_ts,omitempty"`
}

// ReactionItem represents the item a reaction was added to.
type ReactionItem struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// ActionDetail represents a single action in an interactive message payload.
type ActionDetail struct {
	ActionID string `json:"action_id"`
	Type     string `json:"type"`
	Value    string `json:"value,omitempty"`
}

// IDField is a helper for Slack objects that have an "id" field.
type IDField struct {
	ID string `json:"id"`
}

// Message represents a Slack message within an action payload.
type Message struct {
	Text     string `json:"text"`
	TS       string `json:"ts"`
	ThreadTS string `json:"thread_ts,omitempty"`
	User     string `json:"user,omitempty"`
}

// URLVerificationResponse is the response for url_verification challenges.
type URLVerificationResponse struct {
	Challenge string `json:"challenge"`
}

// ThreadMessage represents a single message in a thread, used when bundling
// thread history into SQS messages.
type ThreadMessage struct {
	User string `json:"user"`
	Text string `json:"text"`
	TS   string `json:"ts"`
}

// Event type constants.
const (
	EventTypeURLVerification = "url_verification"
	EventTypeEventCallback   = "event_callback"
	EventTypeReactionAdded   = "reaction_added"
	EventTypeMessageAction   = "message_action"
)
