package pay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostJSONSendsPayloadAndReturnsResponse(t *testing.T) {
	var gotPayload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("content-type = %q, want application/json", contentType)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok"}`))
	}))
	defer server.Close()

	respBody, err := PostJSON("TESTPAY", server.URL, map[string]string{
		"merOrderNo": "WD202605090003",
		"sign":       "abc123",
	})
	if err != nil {
		t.Fatalf("PostJSON() error = %v", err)
	}
	if string(respBody) != `{"code":200,"msg":"ok"}` {
		t.Fatalf("response body = %s", string(respBody))
	}
	if gotPayload["merOrderNo"] != "WD202605090003" {
		t.Fatalf("merOrderNo = %q", gotPayload["merOrderNo"])
	}
	if gotPayload["sign"] != "abc123" {
		t.Fatalf("sign = %q", gotPayload["sign"])
	}
}
