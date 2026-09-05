package config

import "wecheckin/backend/internal/support/notificationstyle"

func ApplyDingTalkTemplate(template notificationstyle.DingTalkTemplate, payload DingTalkWorkNotificationPayload) DingTalkWorkNotificationPayload {
	rendered := notificationstyle.RenderDingTalkTemplate(template, notificationstyle.DingTalkTemplateData{
		Title: payload.Title, Content: payload.Content, URL: payload.URL, PicURL: payload.PicURL,
		SourceName: payload.SourceName, MediaID: payload.MediaID, Duration: payload.Duration,
	})
	messageType := rendered.MessageType
	if messageType == notificationstyle.DingTalkMessageTypeAuto {
		messageType = payload.MessageType
	}
	return DingTalkWorkNotificationPayload{
		MessageType: messageType,
		Title:       rendered.Title, Content: rendered.Content, URL: rendered.URL,
		PicURL: rendered.PicURL, SourceName: rendered.SourceName, MediaID: rendered.MediaID,
		Duration: rendered.Duration, ButtonTitle: rendered.ButtonTitle, HeadColor: rendered.HeadColor,
	}
}
