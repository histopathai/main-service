package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/stretchr/testify/assert"
)

// The handlers are nil on purpose: this test is about which requests the role
// gates let through, not about what happens afterwards. A request that passes
// reaches a nil handler, panics and is answered 500 by gin's recovery — so
// "allowed" is measured as "not 401 and not 403".
func newGateRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter, gin.DefaultErrorWriter = io.Discard, io.Discard
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := NewRouter(&RouterConfig{Logger: logger, RequestTimeout: time.Second},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		middleware.NewAuthMiddleware(logger),
		middleware.NewTimeoutMiddleware(time.Second, logger))
	return r.SetupRoutes()
}

func status(engine *gin.Engine, method, path, role string) int {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-User-ID", "u1")
	if role != "" {
		req.Header.Set("X-User-Role", role)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestRoleGates(t *testing.T) {
	engine := newGateRouter()

	// One write and one read per group that is read-only for data scientists.
	readOnlyGroups := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/workspaces"},
		{http.MethodPut, "/api/v1/workspaces/w1"},
		{http.MethodDelete, "/api/v1/workspaces/w1/soft-delete"},
		{http.MethodPost, "/api/v1/patients"},
		{http.MethodPut, "/api/v1/patients/p1"},
		{http.MethodPost, "/api/v1/images"},
		{http.MethodPut, "/api/v1/images/i1"},
		{http.MethodPost, "/api/v1/annotations"},
		{http.MethodPut, "/api/v1/annotations/a1"},
		{http.MethodDelete, "/api/v1/annotations/a1/soft-delete"},
		{http.MethodDelete, "/api/v1/annotations/soft-delete-many"},
		{http.MethodPost, "/api/v1/annotation-reviews"},
		{http.MethodPut, "/api/v1/annotation-reviews/r1"},
		{http.MethodDelete, "/api/v1/annotation-reviews/r1"},
		{http.MethodPost, "/api/v1/annotation-types"},
		{http.MethodPut, "/api/v1/annotation-types/t1"},
	}
	reads := []string{
		"/api/v1/workspaces", "/api/v1/patients/p1", "/api/v1/images/i1",
		"/api/v1/annotations/image/i1", "/api/v1/annotations/workspace/w1/label-sets",
		"/api/v1/annotation-reviews/annotation/a1",
		"/api/v1/annotation-types", "/api/v1/tissue-masks/image/i1", "/api/v1/proxy/i1/image.dzi",
	}
	tissueWrites := []struct{ method, path string }{
		{http.MethodPut, "/api/v1/tissue-masks/image/i1"},
		{http.MethodPost, "/api/v1/tissue-masks/image/i1/approve"},
		{http.MethodPost, "/api/v1/tissue-masks/image/i1/reject"},
	}

	allowed := func(code int) bool { return code != http.StatusUnauthorized && code != http.StatusForbidden }

	for _, route := range readOnlyGroups {
		if code := status(engine, route.method, route.path, "pathologist"); code == http.StatusNotFound {
			t.Fatalf("%s %s is not a route any more — update this test", route.method, route.path)
		}
		for _, role := range []string{"admin", "pathologist", "user"} {
			assert.True(t, allowed(status(engine, route.method, route.path, role)), "%s %s as %s", route.method, route.path, role)
		}
		for _, role := range []string{"datascientist", "viewer"} {
			assert.Equal(t, http.StatusForbidden, status(engine, route.method, route.path, role), "%s %s as %s", route.method, route.path, role)
		}
	}

	for _, path := range reads {
		for _, role := range []string{"admin", "pathologist", "datascientist", "user", "viewer"} {
			assert.True(t, allowed(status(engine, http.MethodGet, path, role)), "GET %s as %s", path, role)
		}
	}

	// Tissue masks are read-write for every group.
	for _, route := range tissueWrites {
		for _, role := range []string{"admin", "pathologist", "datascientist"} {
			assert.True(t, allowed(status(engine, route.method, route.path, role)), "%s %s as %s", route.method, route.path, role)
		}
	}

	// Blind tests: every group may take one; the results carry the answer key
	// and are for admins only.
	blindTakes := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/blind-tests"},
		{http.MethodGet, "/api/v1/blind-tests/s1"},
		{http.MethodGet, "/api/v1/blind-tests/s1/images/i1"},
		{http.MethodPut, "/api/v1/blind-tests/s1/answers/i1"},
		{http.MethodPut, "/api/v1/blind-tests/s1/notes/i1"},
		{http.MethodPost, "/api/v1/blind-tests/s1/complete"},
	}
	for _, route := range blindTakes {
		for _, role := range []string{"admin", "pathologist", "datascientist"} {
			assert.True(t, allowed(status(engine, route.method, route.path, role)), "%s %s as %s", route.method, route.path, role)
		}
		assert.Equal(t, http.StatusForbidden, status(engine, route.method, route.path, "unassigned"), "%s %s as unassigned", route.method, route.path)
	}
	assert.True(t, allowed(status(engine, http.MethodGet, "/api/v1/blind-tests/s1/results", "admin")), "results as admin")
	for _, role := range []string{"pathologist", "datascientist", "user", "viewer", "unassigned"} {
		assert.Equal(t, http.StatusForbidden, status(engine, http.MethodGet, "/api/v1/blind-tests/s1/results", role), "results as %s", role)
	}

	// Invitation links of a blind test: admins only.
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/blind-tests/s1/invites"},
		{http.MethodGet, "/api/v1/blind-tests/s1/invites"},
		{http.MethodPut, "/api/v1/blind-tests/s1/invites/inv1"},
	} {
		assert.True(t, allowed(status(engine, route.method, route.path, "admin")), "%s %s as admin", route.method, route.path)
		for _, role := range []string{"pathologist", "datascientist", "unassigned"} {
			assert.Equal(t, http.StatusForbidden, status(engine, route.method, route.path, role), "%s %s as %s", route.method, route.path, role)
		}
	}

	// Public invitation routes need no platform user (the handler checks the
	// link and guest tokens); without any identity header they are not refused
	// by the user gates.
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/public/blind-tests/invites/tok"},
		{http.MethodPost, "/api/v1/public/blind-tests/invites/tok/join"},
		{http.MethodPost, "/api/v1/public/blind-tests/invites/tok/resume"},
		{http.MethodGet, "/api/v1/public/blind-tests/invites/tok/test"},
		{http.MethodPut, "/api/v1/public/blind-tests/invites/tok/test/answers/i1"},
		{http.MethodPut, "/api/v1/public/blind-tests/invites/tok/test/notes/i1"},
		{http.MethodPost, "/api/v1/public/blind-tests/invites/tok/test/complete"},
		{http.MethodGet, "/api/v1/public/blind-tests/invites/tok/test/images/i1"},
	} {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		assert.True(t, allowed(w.Code) && w.Code != http.StatusNotFound, "%s %s without a user: %d", route.method, route.path, w.Code)
	}
	// ...and they reach nothing else: results stay behind the admin gate.
	assert.Equal(t, http.StatusNotFound, status(engine, http.MethodGet, "/api/v1/public/blind-tests/s1/results", ""))

	// Not in a group: nothing under /api/v1, reads included.
	for _, role := range []string{"unassigned", "something-else"} {
		for _, path := range reads {
			assert.Equal(t, http.StatusForbidden, status(engine, http.MethodGet, path, role), "GET %s as %s", path, role)
		}
		assert.Equal(t, http.StatusForbidden, status(engine, http.MethodPut, "/api/v1/tissue-masks/image/i1", role), "tissue write as %s", role)
	}
	assert.Equal(t, http.StatusUnauthorized, status(engine, http.MethodGet, "/api/v1/workspaces", ""), "no role header")

	// Health checks stay open.
	assert.Equal(t, http.StatusOK, status(engine, http.MethodGet, "/health", ""))
}
