package handlers

import (
	"errors"
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/internal/middleware"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/internal/db"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func ProductsRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Post("/", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		uuid, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		product := requests.CreateProduct{}
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		result, err := db.InsertProduct(&product, uuid, r.Context())
		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "The product was created successfully",
			Data: result,
		}

		resp.WriteMessage(w)
	})
}


