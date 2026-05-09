package pay

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// PostJSON calls a third-party JSON API and logs the full request/response pair.
func PostJSON(provider string, url string, payload map[string]string) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] third-party request marshal failed url=%s err=%v", provider, url, err)
		return nil, err
	}
	log.Printf("[%s] third-party request url=%s params=%s", provider, url, string(body))

	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		log.Printf("[%s] third-party request failed url=%s err=%v", provider, url, err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	log.Printf("[%s] third-party response url=%s status=%d body=%s", provider, url, resp.StatusCode, string(respBody))
	if readErr != nil {
		log.Printf("[%s] third-party response read failed url=%s status=%d err=%v", provider, url, resp.StatusCode, readErr)
		return respBody, readErr
	}
	return respBody, nil
}
