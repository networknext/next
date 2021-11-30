package looker

type LookerSave struct {
	SessionID    int64  `json:"session_id"`
	CustomerCode string `json:"customer_code"`
}

type LookerWebhookAttachment struct {
	Data []LookerSave `json:"data"` // TODO: Potentially break this out into a bunch of different webhook attachment types
}

type LookerWebhookScheduledPlan struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	SchedulePlanID int    `json:"schedule_plan_id"`
	Type           string `json:"type"`
}

type LookerWebhookPayload struct {
	Attachment    LookerWebhookAttachment    `json:"attachment"`
	ScheduledPlan LookerWebhookScheduledPlan `json:"scheduled_plan"`
}
