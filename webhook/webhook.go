package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SendWebhookNotification(webhookURL string, data map[string]interface{}) error {
	if webhookURL == "" {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
