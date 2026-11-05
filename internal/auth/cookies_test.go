package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAuthCookie(t *testing.T) {
	behindProxy := httptest.NewRequest(http.MethodPost, "/login", nil)
	behindProxy.Header.Set("X-Forwarded-Proto", "https")
	c := newAuthCookie(behindProxy, "accessToken", "v", 15*time.Minute)
	if !c.Secure || c.SameSite != http.SameSiteNoneMode || !c.HttpOnly || c.MaxAge != 900 {
		t.Fatalf("https cookie = %+v", c)
	}

	plain := httptest.NewRequest(http.MethodPost, "/login", nil)
	c = newAuthCookie(plain, "accessToken", "v", time.Minute)
	if c.Secure || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("http cookie = %+v", c)
	}

	c = newAuthCookie(behindProxy, "accessToken", "", 0)
	if c.MaxAge != -1 || !c.Secure {
		t.Fatalf("clear cookie = %+v", c)
	}
}

func TestRequireRole(t *testing.T) {
	h := RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for role, want := range map[string]int{"admin": 200, "fleet_manager": 403, "": 403} {
		req := httptest.NewRequest(http.MethodGet, "/admin/user/list", nil)
		req = req.WithContext(context.WithValue(req.Context(), RoleContextKey, role))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("role %q: got %d, want %d", role, rec.Code, want)
		}
	}
}
