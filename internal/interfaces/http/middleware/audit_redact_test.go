package middleware

import (
	"encoding/json"
	"testing"
)

func TestRedactJSON_PasswordField(t *testing.T) {
	body := []byte(`{"email":"a@b.com","password":"secret123"}`)
	got := redactJSON(body)

	var data map[string]string
	if err := json.Unmarshal(got, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data["password"] != "[REDACTED]" {
		t.Fatalf("password = %q, want [REDACTED]", data["password"])
	}
	if data["email"] != "a@b.com" {
		t.Fatalf("email = %q, want a@b.com", data["email"])
	}
}

func TestRedactJSON_NonJSON(t *testing.T) {
	got := redactJSON([]byte("not-json"))
	if string(got) != `"[REDACTED]"` {
		t.Fatalf("got %q, want [REDACTED] string", string(got))
	}
}
