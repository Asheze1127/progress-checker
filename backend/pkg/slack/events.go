package slack

// ThreadMessage represents a single message in a thread, used when bundling
// thread history into SQS messages.
type ThreadMessage struct {
  User string `json:"user"`
  Text string `json:"text"`
  TS   string `json:"ts"`
}

// MessageActionPayload represents a Slack message_action (shortcut) payload.
// This type is needed because the slack-go library does not provide a dedicated
// type for message_action payloads in the slackevents package.
type MessageActionPayload struct {
  Type       string `json:"type"`
  CallbackID string `json:"callback_id"`
  TriggerID  string `json:"trigger_id"`
  User       struct {
    ID string `json:"id"`
  } `json:"user"`
  Channel struct {
    ID string `json:"id"`
  } `json:"channel"`
  Message struct {
    Text     string `json:"text"`
    TS       string `json:"ts"`
    ThreadTS string `json:"thread_ts"`
  } `json:"message"`
}

// Event type constant for message_action, which is not provided by the
// slack-go slackevents package.
const EventTypeMessageAction = "message_action"
