package handlers

import (
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/internal/middleware"
	"github.com/go-chi/chi"
)

func ProductsRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Post("/", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode("products ping")
	})
}


