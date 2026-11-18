package email

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
)

type resendMailer struct {
	apiKey string
	from   *mail.Address
}

func (m *resendMailer) Name() string { return "resend" }

func (m *resendMailer) Send(to, subject, htmlBody, _ string) error {
	payload := map[string]any{
		"from":    m.from.String(),
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return fmt.Errorf("resend API error: %d %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}
