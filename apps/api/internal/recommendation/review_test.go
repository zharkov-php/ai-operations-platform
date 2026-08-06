package recommendation

import (
	"errors"
	"testing"
)

func TestNextReviewStatus(t *testing.T) {
	tests := []struct {
		name, current, action, want string
		wantErr                     error
	}{
		{name: "accept new", current: "new", action: "accept", want: "accepted"},
		{name: "reject review", current: "under_review", action: "reject", want: "rejected"},
		{name: "duplicate accepted", current: "accepted", action: "accept", wantErr: ErrInvalidTransition},
		{name: "duplicate rejected", current: "rejected", action: "reject", wantErr: ErrInvalidTransition},
		{name: "unknown action", current: "new", action: "apply", wantErr: ErrInvalidTransition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NextReviewStatus(test.current, test.action)
			if got != test.want || !errors.Is(err, test.wantErr) {
				t.Fatalf("status=%q err=%v", got, err)
			}
		})
	}
}
