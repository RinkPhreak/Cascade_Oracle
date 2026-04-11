package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cascade/internal/application/port"
)

type smsRuClient struct {
	apiID      string
	httpClient *http.Client
}

func NewSMSRuClient(apiID string, timeout time.Duration) port.SMSClient {
	return &smsRuClient{
		apiID: apiID,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *smsRuClient) isMockMode() bool {
	return os.Getenv("MESSENGER_MODE") == "mock"
}

func (c *smsRuClient) Send(ctx context.Context, phone string, message string) (int, error) {
	start := time.Now()

	if c.isMockMode() {
		return 150, nil
	}

	url := fmt.Sprintf("https://sms.ru/sms/send?api_id=%s&to=%s&msg=%s&json=1", c.apiID, phone, message)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return int(time.Since(start).Milliseconds()), fmt.Errorf("sms.ru returned HTTP %d", resp.StatusCode)
	}

	var parsed map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &parsed); err != nil {
		return int(time.Since(start).Milliseconds()), err
	}

	statusCode, ok := parsed["status_code"].(float64)
	if !ok || statusCode != 100 {
		return int(time.Since(start).Milliseconds()), fmt.Errorf("sms.ru API error: %v", parsed)
	}

	return int(time.Since(start).Milliseconds()), nil
}

func (c *smsRuClient) GetBalance(ctx context.Context) (float64, error) {
	if c.isMockMode() {
		return 1000.0, nil
	}

	url := fmt.Sprintf("https://sms.ru/my/balance?api_id=%s&json=1", c.apiID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var parsed map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, err
	}
	
	balance, ok := parsed["balance"].(float64)
	if !ok {
		return 0, fmt.Errorf("sms.ru API error: bad balance response")
	}

	return balance, nil
}
