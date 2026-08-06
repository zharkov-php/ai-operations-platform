package notification

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid notification")

type Event struct{ Type, ResourceID, Title, Body, DeepLink, DedupeKey string }
type Preferences struct {
	InApp      bool `json:"in_app"`
	MobilePush bool `json:"mobile_push"`
}
type Delivery struct {
	Attempts    int
	NextAttempt time.Time
	Status      string
}
type Channel interface {
	Deliver(context.Context, Event) error
}
type Dispatcher struct{ channels []Channel }

func NewDispatcher(channels ...Channel) Dispatcher { return Dispatcher{channels: channels} }
func (d Dispatcher) Deliver(ctx context.Context, event Event) error {
	if err := Validate(event); err != nil {
		return err
	}
	for _, channel := range d.channels {
		if err := channel.Deliver(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

type ConsoleChannel struct{ Log func(string, ...any) }

func (c ConsoleChannel) Deliver(_ context.Context, event Event) error {
	if c.Log != nil {
		c.Log("development notification", "event_type", event.Type, "deep_link", event.DeepLink)
	}
	return nil
}

var eventTypes = map[string]bool{"budget_threshold_reached": true, "high_priority_recommendation": true, "experiment_guardrail_violation": true, "automatic_rollback": true, "evaluation_completed": true, "experiment_verified": true}

func Validate(event Event) error {
	if !eventTypes[event.Type] || strings.TrimSpace(event.Title) == "" || strings.TrimSpace(event.DedupeKey) == "" || len(event.Body) > 500 {
		return ErrInvalid
	}
	u, err := url.Parse(event.DeepLink)
	if err != nil || u.Scheme != "ai-execution-advisor" || u.Host != "open" || strings.ContainsAny(event.DeepLink, "\r\n") {
		return ErrInvalid
	}
	if strings.Contains(strings.ToLower(event.Body), "bearer ") || strings.Contains(strings.ToLower(event.Body), "api_key") {
		return ErrInvalid
	}
	return nil
}

func Retry(attempts int, now time.Time) Delivery {
	if attempts >= 5 {
		return Delivery{Attempts: attempts, Status: "failed"}
	}
	delay := time.Minute * time.Duration(1<<attempts)
	if delay > time.Hour {
		delay = time.Hour
	}
	return Delivery{Attempts: attempts + 1, NextAttempt: now.UTC().Add(delay), Status: "pending"}
}

func DeepLink(resource string, id string) string {
	return fmt.Sprintf("ai-execution-advisor://open/%s/%s", url.PathEscape(resource), url.PathEscape(id))
}
