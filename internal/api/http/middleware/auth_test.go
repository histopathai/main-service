package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	am := NewAuthMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)))

	cases := map[string]int{"admin": http.StatusOK, "ADMIN": http.StatusOK, "user": http.StatusForbidden, "": http.StatusUnauthorized}
	for role, want := range cases {
		r := gin.New()
		r.Use(am.RequireAuth(), am.RequireRole("admin"))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", "u1")
		if role != "" {
			req.Header.Set("X-User-Role", role)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, want, w.Code, "role %q", role)
	}
}

func TestRequireAuthNormalizesLegacyRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	am := NewAuthMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)))

	cases := map[string]string{
		"user":          RolePathologist,
		"viewer":        RoleDatascientist,
		"ADMIN":         RoleAdmin,
		" Pathologist ": RolePathologist,
		"datascientist": RoleDatascientist,
		"unassigned":    "unassigned",
	}
	for header, want := range cases {
		var got string
		r := gin.New()
		r.Use(am.RequireAuth())
		r.GET("/", func(c *gin.Context) { got, _ = GetAuthenticatedUserRole(c) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", "u1")
		req.Header.Set("X-User-Role", header)
		r.ServeHTTP(httptest.NewRecorder(), req)
		assert.Equal(t, want, got, "header %q", header)
	}
}

func TestDenyWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	am := NewAuthMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)))

	reads := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	writes := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	serve := func(method, role string, withAuth bool) int {
		r := gin.New()
		if withAuth {
			r.Use(am.RequireAuth())
		}
		r.Use(am.DenyWrites(RoleDatascientist))
		r.Handle(method, "/", func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(method, "/", nil)
		req.Header.Set("X-User-ID", "u1")
		req.Header.Set("X-User-Role", role)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	for _, m := range reads {
		// "viewer" is the legacy name of datascientist.
		for _, role := range []string{RoleAdmin, RolePathologist, RoleDatascientist, "viewer"} {
			assert.Equal(t, http.StatusOK, serve(m, role, true), "%s as %s", m, role)
		}
	}
	for _, m := range writes {
		assert.Equal(t, http.StatusOK, serve(m, RoleAdmin, true), "%s as admin", m)
		assert.Equal(t, http.StatusOK, serve(m, RolePathologist, true), "%s as pathologist", m)
		assert.Equal(t, http.StatusOK, serve(m, "user", true), "%s as legacy user", m)
		assert.Equal(t, http.StatusForbidden, serve(m, RoleDatascientist, true), "%s as datascientist", m)
		assert.Equal(t, http.StatusForbidden, serve(m, "DataScientist", true), "%s as DataScientist", m)
		assert.Equal(t, http.StatusForbidden, serve(m, "viewer", true), "%s as legacy viewer", m)
		// Without RequireAuth in front there is no role to judge: fail closed.
		assert.Equal(t, http.StatusForbidden, serve(m, RoleAdmin, false), "%s without RequireAuth", m)
	}
}
