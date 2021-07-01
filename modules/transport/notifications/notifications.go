package notifications

import (
	"fmt"
	"time"
)

type NotificationPriorty int32
type NotificationType int32

const (
	// TODO: Move these somewhere else like the jsonrpc error codes and use them for something
	DEFAULT_PRIORITY NotificationPriorty = 0
	INFO_PRIORITY    NotificationPriorty = 1
	WARNING_PRIORITY NotificationPriorty = 2
	URGENT_PRIORITY  NotificationPriorty = 3
	// TODO: Move these somewhere else like the jsonrpc error codes and actually use them for something
	NOTIFICATION_SYSTEM        NotificationType = 0
	NOTIFICATION_RELEASE_NOTES NotificationType = 1
	NOTIFICATION_ANALYTICS     NotificationType = 2
	NOTIFICATION_INVOICE       NotificationType = 3
)

type GistEmbed struct {
	CSSURL    string
	EmbedHTML string
}

type ReleaseNotesNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Priority  NotificationPriorty `json:"priority"`
	CSSURL    string              `json:"css_url"`
	EmbedHTML string              `json:"embed_html"`
}

func NewReleaseNotesNotification() ReleaseNotesNotification {
	return ReleaseNotesNotification{
		Type:     NOTIFICATION_RELEASE_NOTES,
		Priority: DEFAULT_PRIORITY,
	}
}

type SystemNotification struct {
	Type     NotificationType    `json:"type"`
	Title    string              `json:"title"`
	Message  string              `json:"message"`
	Priority NotificationPriorty `json:"priority"`
}

func NewSystemNotifications() SystemNotification {
	return SystemNotification{
		Type:     NOTIFICATION_SYSTEM,
		Priority: DEFAULT_PRIORITY,
	}
}

type AnalyticsNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Message   string              `json:"message"`
	Priority  NotificationPriorty `json:"priority"`
	LookerURL string              `json:"looker_url"`
}

func NewAnalyticsNotification() AnalyticsNotification {
	return AnalyticsNotification{
		Type:     NOTIFICATION_ANALYTICS,
		Priority: DEFAULT_PRIORITY,
	}
}

type InvoiceNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Message   string              `json:"message"`
	Priority  NotificationPriorty `json:"priority"`
	InvoiceID string              `json:"invoice_id"`
}

func NewInvoiceNotification() InvoiceNotification {
	return InvoiceNotification{
		Title:    fmt.Sprintf("Invoice for the month of %s", time.Now().Month()),
		Message:  "Your invoice is ready and can be viewed in the \"Invoicing\" tab",
		Type:     NOTIFICATION_INVOICE,
		Priority: DEFAULT_PRIORITY,
	}
}
