package httphandler

import (
	"net/http"

	"github.com/AmirSoleimani/openberth/apps/server/internal/service"
)

func (h *Handlers) SecretSet(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	var req struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
		Global      bool   `json:"global"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		jsonErr(w, 400, "Invalid JSON: "+err.Error())
		return
	}

	result, err := h.svc.SecretSet(user, req.Name, req.Value, req.Description, req.Global)
	if err != nil {
		writeErr(w, err)
		return
	}

	jsonResp(w, 200, result)
}

func (h *Handlers) SecretList(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	secrets, err := h.svc.SecretList(user)
	if err != nil {
		writeErr(w, err)
		return
	}
	if secrets == nil {
		secrets = []service.SecretMeta{} // ensure JSON array not null
	}

	jsonResp(w, 200, map[string]interface{}{"secrets": secrets})
}

func (h *Handlers) SecretDelete(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	name := r.PathValue("name")
	// Check for ?global=true query param
	global := r.URL.Query().Get("global") == "true"

	if err := h.svc.SecretDelete(user, name, global); err != nil {
		writeErr(w, err)
		return
	}

	jsonResp(w, 200, map[string]string{"name": name, "status": "deleted"})
}
