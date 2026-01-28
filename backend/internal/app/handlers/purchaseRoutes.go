package handlers

import (
	"time"
	"errors"
	"strconv"
	"strings"
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
	
	router.Get("/", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := 1

		pageStr := r.URL.Query().Get("page")
		if pageStr != "" {
			parsedPage, err := strconv.Atoi(pageStr)

			if err != nil || parsedPage <= 0 {
				tools.BadRequestErrorHandler(w, errors.New("Invalid Page number"))
				return
			}

			page = parsedPage
		}
		
		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		tickets, err := db.FetchTickets(page, userID, r.Context())
		if err != nil {
			tools.InternalServerErrorHandler(w, nil)
			return
		}


		resp := tools.Message {
			Message: "List of purchases",
			Data: tickets,
		}

		resp.WriteMessage(w)
	})

	router.Get("/{purchase}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		purchaseID := chi.URLParam(r, "purchase")
		if strings.TrimSpace(purchaseID) == "" {
			tools.BadRequestErrorHandler(w, errors.New("Invalid purchase id"))
			return
		}

		purchases, err := db.FetchPurchase(purchaseID, userID, r.Context())
		if err != nil {
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if len(purchases) == 0 {
			tools.NotFoundErrorHandler(w, "Purchase not found")
			return
		}
		
		resp := tools.Message {
			Message: "Purchase details",
			Data: purchases,
		}

		resp.WriteMessage(w)
	})


	router.Delete("/{purchase}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		purchaseID := chi.URLParam(r, "purchase")
		if strings.TrimSpace(purchaseID) == "" {
			tools.BadRequestErrorHandler(w, errors.New("Invalid purchase id"))
			return
		}

		purchases, err := db.FetchPurchase(purchaseID, userID, r.Context())
		if err != nil {
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if len(purchases) == 0 {
			tools.NotFoundErrorHandler(w, "Purchase not found")
			return
		}

		if time.Since(purchases[0].CreatedAt) > time.Hour {
			tools.BadRequestErrorHandler(w, errors.New("The purchase its to old to delete"))
			return
		}

		if err := db.DeletePurchase(purchases, r.Context()); err != nil {
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "Purchase deleted successfully",
			Data: true,
		}

		resp.WriteMessage(w)
	})
}
