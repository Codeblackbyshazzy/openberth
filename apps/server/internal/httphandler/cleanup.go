package httphandler

import (
	"log"
	"net/http"
)

// Cleanup removes expired deployments, expired OAuth data, and expired sessions.
// Intended for local invocation only (berth-admin cleanup). Reject any request
// that arrived via the public reverse proxy — Caddy always adds the
// X-Forwarded-* headers when proxying, so their presence is a reliable
// signal that the request did not originate from localhost.
func (h *Handlers) Cleanup(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Forwarded-Host") != "" || r.Header.Get("X-Forwarded-Proto") != "" {
		http.NotFound(w, r)
		return
	}
	count := h.svc.RunCleanup()
	jsonResp(w, 200, map[string]int{"cleaned": count})
}

// DestroyAllDeployments removes all deployments belonging to the authenticated user.
func (h *Handlers) DestroyAllDeployments(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	deploys, _ := h.svc.Store.ListDeployments(user.ID)
	count := 0
	for _, d := range deploys {
		h.svc.DestroyFull(&d)
		count++
	}

	log.Printf("[destroy-all] %d deployments | user=%s", count, user.Name)
	jsonResp(w, 200, map[string]int{"destroyed": count})
}
