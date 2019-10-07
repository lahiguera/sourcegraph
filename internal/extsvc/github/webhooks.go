package github

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type Webhook struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	Config WebhookConfig `json:"config"`
	Events []string      `json:"events"`
}

type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"` // json or form
	Secret      string `json:"secret"`
}

func (c *Client) ListOrgWebhooks(ctx context.Context, org string) (ws []Webhook, _ error) {
	return ws, c.requestGet(ctx, "", "orgs/"+org+"/hooks", &ws)
}

func (c *Client) CreateOrgWebhook(ctx context.Context, org string, w *Webhook) error {
	w.Name = "web"

	body, err := json.Marshal(w)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "orgs/"+org+"/hooks", bytes.NewReader(body))
	if err != nil {
		return err
	}

	return c.do(ctx, "", req, w)
}
