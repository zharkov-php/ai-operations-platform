package notification

import (
	"context"
	"errors"
	"testing"
	"time"
)

type recordingChannel struct{ event Event }

func (c *recordingChannel) Deliver(_ context.Context, event Event) error { c.event = event; return nil }

func TestValidateSafePayload(t *testing.T) {
	good := Event{Type: "automatic_rollback", Title: "Experiment rolled back", Body: "A cost guardrail was exceeded.", DeepLink: DeepLink("experiments", "abc"), DedupeKey: "rollback:abc"}
	if err := Validate(good); err != nil {
		t.Fatal(err)
	}
	bad := good
	bad.Body = "Bearer secret"
	if !errors.Is(Validate(bad), ErrInvalid) {
		t.Fatal("sensitive payload accepted")
	}
}
func TestRetryPolicy(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := Retry(2, now)
	if got.Attempts != 3 || got.NextAttempt.Sub(now) != 4*time.Minute || got.Status != "pending" {
		t.Fatalf("unexpected retry: %+v", got)
	}
	if Retry(5, now).Status != "failed" {
		t.Fatal("exhausted delivery must fail")
	}
}
func TestDeepLink(t *testing.T) {
	event := Event{Type: "evaluation_completed", Title: "Evaluation completed", Body: "Results are ready.", DeepLink: DeepLink("evaluations", "case one"), DedupeKey: "eval:1"}
	if Validate(event) != nil {
		t.Fatal("generated deep link invalid")
	}
}
func TestDispatcherExcludesSensitivePayloads(t *testing.T) {
	channel := &recordingChannel{}
	dispatcher := NewDispatcher(channel)
	event := Event{Type: "evaluation_completed", Title: "Complete", Body: "Bearer secret", DeepLink: DeepLink("evaluations", "one"), DedupeKey: "one"}
	if !errors.Is(dispatcher.Deliver(context.Background(), event), ErrInvalid) || channel.event.Type != "" {
		t.Fatal("unsafe event reached channel")
	}
}
