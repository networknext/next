package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const DefaultSlackTimeout = 5 * time.Second

type SlackClient struct {
	WebHookUrl string
	UserName   string
	Channel    string
	TimeOut    time.Duration
}

type SimpleSlackRequest struct {
	Text string
}

type SlackJobNotification struct {
	Color   string
	Details string
	Text    string
}

type SlackMessage struct {
	Username    string       `json:"username,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color         string `json:"color,omitempty"`
	Fallback      string `json:"fallback,omitempty"`
	CallbackID    string `json:"callback_id,omitempty"`
	ID            int    `json:"id,omitempty"`
	AuthorID      string `json:"author_id,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`
	Title         string `json:"title,omitempty"`
	TitleLink     string `json:"title_link,omitempty"`
	Pretext       string `json:"pretext,omitempty"`
	Text          string `json:"text,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	ThumbURL      string `json:"thumb_url,omitempty"`
	// Fields and actions are not defined.
	MarkdownIn []string    `json:"mrkdwn_in,omitempty"`
	Ts         json.Number `json:"ts,omitempty"`
}

// SendSlackNotification will post to an 'Incoming Webook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (sc SlackClient) SendSlackNotification(sr SimpleSlackRequest) error {
	slackRequest := SlackMessage{
		Text:     sr.Text,
		Username: sc.UserName,
		Channel:  sc.Channel,
	}
	return sc.sendHttpRequest(slackRequest)
}

func (sc SlackClient) SendJobNotification(job SlackJobNotification) error {
	attachment := Attachment{
		Color: job.Color,
		Text:  job.Details,
		Ts:    json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	slackRequest := SlackMessage{
		Text:        job.Text,
		Username:    sc.UserName,
		Channel:     sc.Channel,
		Attachments: []Attachment{attachment},
	}
	return sc.sendHttpRequest(slackRequest)
}

func (sc SlackClient) SendError(message string) (err error) {
	return sc.sendFunc("danger", message)
}

func (sc SlackClient) SendInfo(message string) (err error) {
	return sc.sendFunc("good", message)
}

func (sc SlackClient) SendWarning(message string) (err error) {
	return sc.sendFunc("warning", message)
}

func (sc SlackClient) sendFunc(color string, message string) error {
	sjn := SlackJobNotification{
		Color:   color,
		Details: message,
	}
	return sc.SendJobNotification(sjn)
}
func (sc SlackClient) sendHttpRequest(slackRequest SlackMessage) error {
	slackBody, _ := json.Marshal(slackRequest)
	req, err := http.NewRequest(http.MethodPost, sc.WebHookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if sc.TimeOut == 0 {
		sc.TimeOut = DefaultSlackTimeout
	}
	client := &http.Client{Timeout: sc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
