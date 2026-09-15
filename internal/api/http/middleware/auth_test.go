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
