package redaction

import (
	"strings"
	"testing"
)

func TestRedactsSecretsAndContacts(t *testing.T) {
	input := "Authorization: Bearer secret.token email owner@example.test phone +1 555 123 4567 password=hunter2"
	output := Text(input, true)
	for _, secret := range []string{"secret.token", "owner@example.test", "555 123", "hunter2"} {
		if strings.Contains(output, secret) {
			t.Fatalf("redaction leaked %q: %s", secret, output)
		}
	}
}
func TestRedactsNestedMetadata(t *testing.T) {
	clean := Metadata(map[string]any{"authorization": "Bearer secret", "nested": map[string]any{"api_key": "secret"}, "count": 2}, false)
	if clean["authorization"] != replacement || clean["nested"].(map[string]any)["api_key"] != replacement || clean["count"] != 2 {
		t.Fatalf("metadata=%+v", clean)
	}
}
func TestPreviewIsBounded(t *testing.T) {
	preview := Preview(strings.Repeat("a", 100), 20)
	if len(preview) > 23 {
		t.Fatalf("preview length=%d", len(preview))
	}
}
func TestPreviewPreservesUnicode(t *testing.T) {
	preview := Preview(strings.Repeat("🙂", 10), 3)
	if preview != "🙂🙂🙂…" {
		t.Fatalf("preview=%q", preview)
	}
}
