package handlers

import (
	"errors"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"path/filepath"

	"github.com/OscarVillanueva/goapi/internal/app/internal/middleware"
	"github.com/OscarVillanueva/goapi/internal/app/models/parameters"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/internal/db"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/platform"
	
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func ProductsRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Get("/", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := 1
		onlyAvailable := false

		userID, ok := r.Context().Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		pageStr := r.URL.Query().Get("page")
		if pageStr != "" {
			parsedPage, err := strconv.Atoi(pageStr)

			if err != nil || parsedPage <= 0 {
				tools.BadRequestErrorHandler(w, errors.New("Invalid Page number"))
				return
			}

			page = parsedPage
		}

		availableStr := r.URL.Query().Get("available")
		if availableStr != "" {
			parsedAvailable, err := strconv.ParseBool(availableStr)

			if err != nil {
				tools.BadRequestErrorHandler(w, errors.New("Invalid available value"))
				return
			}

			onlyAvailable = parsedAvailable
		}

		search := r.URL.Query().Get("search")

		products, err := db.GetProducts(parameters.GetProductsParams{
			User: userID,
			Page: page,
			Context: r.Context(),
			OnlyAvailable: onlyAvailable,
			SearchName: search,
		})
		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "Products",
			Data: products,
		}

		resp.WriteMessage(w)
	})

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

	router.Put("/{product}/image", func (w http.ResponseWriter, r *http.Request)  {
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

		_, err := db.GetProduct(userID, productID, r.Context())
		if err != nil {
			if errors.Is(err, db.ErrProductNotFound){
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		// << 20 is a short of 2^20 -> 5 * 1MB
		r.ParseMultipartForm(5 << 20)

		file, handler, err := r.FormFile("image")
		if err != nil {
			tools.BadRequestErrorHandler(w, errors.New("We couldn't get the image"))
			return
		}

		defer file.Close()
		
		contentType := handler.Header.Get("Content-Type")
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			tools.BadRequestErrorHandler(w, errors.New("Only jpeg, png, and webp images are allowed"))
			return
		}

		ext := filepath.Ext(handler.Filename)
		objectName := productID + ext

		path, putErr := platform.PutImage(objectName, file, handler.Size, contentType, r.Context())
		if putErr != nil {
			msg := "We couldn't save your image"
			tools.InternalServerErrorHandler(w, &msg)
			return
		}

		if updateErr := db.UpdateProductImage(productID, path, userID, r.Context()); updateErr != nil {
			if errors.Is(updateErr, db.ErrProductNotFound){
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			// Real DB error
			log.Error("Failed to update the product: ", updateErr)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "The image was saved successfully",
			Data: path,
		}

		resp.WriteMessage(w)
	})
}


