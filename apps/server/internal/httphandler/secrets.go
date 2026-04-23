package httphandler

import (
	"net/http"

	"github.com/AmirSoleimani/openberth/apps/server/internal/service"
	"github.com/AmirSoleimani/openberth/apps/server/internal/store"
)

type secretSetRequest struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Global      bool   `json:"global"`
}

func (h *Handlers) SecretSet(w http.ResponseWriter, r *http.Request) {
	authedJSON(h, w, r, func(user *store.User, req secretSetRequest) (any, error) {
		return h.svc.SecretSet(user, req.Name, req.Value, req.Description, req.Global)
	})
}

func (h *Handlers) SecretList(w http.ResponseWriter, r *http.Request) {
	authed(h, w, r, func(user *store.User) (any, error) {
		secrets, err := h.svc.SecretList(user)
		if err != nil {
			return nil, err
		}
		if secrets == nil {
			secrets = []service.SecretMeta{} // ensure JSON array, not null
		}
		return map[string]any{"secrets": secrets}, nil
	})
}

func (h *Handlers) SecretDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	global := r.URL.Query().Get("global") == "true"

	authed(h, w, r, func(user *store.User) (any, error) {
		if err := h.svc.SecretDelete(user, name, global); err != nil {
			return nil, err
		}
		return map[string]string{"name": name, "status": "deleted"}, nil
	})
}
