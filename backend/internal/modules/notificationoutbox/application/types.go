package application

type InternalRecipient struct {
	NotifyAdmin bool   `json:"notifyAdmin"`
	UserIDs     []uint `json:"userIds,omitempty"`
}

type WebhookRecipient struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type MessagePayload struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	SourceType string `json:"sourceType,omitempty"`
	SourceID   string `json:"sourceId,omitempty"`
}
