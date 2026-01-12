package handlers

import (
	"errors"
	"strings"
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
		if !ok || uuid == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		var product requests.CreateProduct
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := product.Validate(); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, err)
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

	router.Delete("/{product}", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		productID := chi.URLParam(r, "product")
		if strings.TrimSpace(productID) == "" {
			tools.BadRequestErrorHandler(w, errors.New("Invalid product id"))
			return
		}

		if err := db.DeleteProduct(productID, userID, r.Context()); err != nil {
			if err.Error() == "Record Not found" {
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			// Real DB error
			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "The product was deleted successfully",
			Data: true,
		}

		resp.WriteMessage(w)
	})

	router.Put("/{product}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		productID := chi.URLParam(r, "product")
		if strings.TrimSpace(productID) == "" {
			tools.BadRequestErrorHandler(w, errors.New("Invalid product id"))
			return
		}

		var product requests.CreateProduct
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := product.Validate(); err != nil {
			tools.BadRequestErrorHandler(w, err)
			return
		}

		if err := db.UpdateProduct(productID, &product, userID, r.Context()); err != nil {
			if errors.Is(err, db.ErrProductNotFound){
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			// Real DB error
			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "The product was updated successfully",
			Data: true,
		}

		resp.WriteMessage(w)

	})
}


