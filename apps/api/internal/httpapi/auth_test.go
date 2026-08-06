package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
)

func TestRoleAuthorizationAndTenantClaims(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(claimsKey{}).(auth.Claims)
		if claims.OrganizationID != "org-a" {
			t.Fatal("tenant claim changed")
		}
		w.WriteHeader(204)
	})
	for _, test := range []struct {
		role string
		want int
	}{{"owner", 204}, {"viewer", 403}} {
		t.Run(test.role, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/", nil)
			request = request.WithContext(context.WithValue(request.Context(), claimsKey{}, auth.Claims{OrganizationID: "org-a", Role: test.role}))
			recorder := httptest.NewRecorder()
			requireRoles("owner", "admin")(next).ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("status=%d want=%d", recorder.Code, test.want)
			}
		})
	}
}
