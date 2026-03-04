package service

type SlashResponse struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text,omitempty"`
}

type ProgressModalMetadata struct {
	ChannelID string `json:"channel_id"`
	ThreadTS  string `json:"thread_ts"`
}

const (
	ProgressOpenModalActionID = "progress_open_modal"
	ProgressModalCallbackID   = "progress_modal_submit"

	ProgressPhaseBlockID  = "progress_phase_block"
	ProgressPhaseActionID = "progress_phase_action"

	ProgressStatusBlockID  = "progress_status_block"
	ProgressStatusActionID = "progress_status_action"

	ProgressNoteBlockID  = "progress_note_block"
	ProgressNoteActionID = "progress_note_action"
)
