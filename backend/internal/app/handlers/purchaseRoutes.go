package handlers

import (
	"errors"
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/app/internal/db"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/internal/middleware"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func PurchaseRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Post("/", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		var ticket []requests.CreatePurchase
		if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		purchaseID, err := db.BatchPurchase(ticket, userID, r.Context())

		if err != nil {
			log.Error(err)
			var stockErr *db.ErrInsufficientStock
			if errors.As(err, &stockErr) {
				tools.BadRequestErrorHandler(w, errors.New(stockErr.Error()))
				return
			}

			if err.Error() == "quantity must be greater than zero" {
				tools.BadRequestErrorHandler(w, err)
				return
			}

			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "Purchase created successfully",
			Data: purchaseID,
		}

		resp.WriteMessage(w)
	})
	
}
