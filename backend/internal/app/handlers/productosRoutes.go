package handlers

import (
	"fmt"
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
	
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

const ProductsRouterName = "products-router"

func ProductsRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Get("/", func (w http.ResponseWriter, r *http.Request) {
		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.GET./", ProductsRouterName))
		defer span.End()

		w.Header().Set("Content-Type", "application/json")
		page := 1
		onlyAvailable := false

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)
		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.SetAttributes(
				attribute.String("uuid", userID),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		pageStr := r.URL.Query().Get("page")

		span.SetAttributes(
			attribute.String("PageStr", pageStr),
		)

		if pageStr != "" {
			parsedPage, err := strconv.Atoi(pageStr)

			if err != nil || parsedPage <= 0 {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				tools.BadRequestErrorHandler(w, errors.New("Invalid Page number"))
				return
			}

			page = parsedPage
		}

		availableStr := r.URL.Query().Get("available")

		span.SetAttributes(
			attribute.String("Available", availableStr),
		)

		if availableStr != "" {
			parsedAvailable, err := strconv.ParseBool(availableStr)

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				tools.BadRequestErrorHandler(w, errors.New("Invalid available value"))
				return
			}

			onlyAvailable = parsedAvailable
		}

		search := r.URL.Query().Get("search")

		span.SetAttributes(
			attribute.String("Search", search),
		)

		products, err := db.GetProducts(parameters.GetProductsParams{
			User: userID,
			Page: page,
			Context: ctx,
			OnlyAvailable: onlyAvailable,
			SearchName: search,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Fetch products successfully")

		resp := tools.Message {
			Message: "Products",
			Data: products,
		}

		resp.WriteMessage(w)
	})

	router.Post("/", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.POST./", ProductsRouterName))
		defer span.End()

		uuid, ok := ctx.Value(middleware.UserUUIDKey).(string)
		if !ok || uuid == ""{
			err := errors.New("Missing user uuid")
			span.SetAttributes(
				attribute.String("uuid", uuid),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		var product requests.CreateProduct
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := product.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		result, err := db.InsertProduct(&product, uuid, ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Create Product successfully")

		resp := tools.Message {
			Message: "The product was created successfully",
			Data: result,
		}

		resp.WriteMessage(w)
	})

	router.Delete("/{product}", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.DELETE./", AuthRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)
		
		span.SetAttributes(
			attribute.String("UserID", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		productID := chi.URLParam(r, "product")

		span.SetAttributes(
			attribute.String("ProductID", productID),
		)

		if strings.TrimSpace(productID) == "" {
			err := errors.New("Invalid product id")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		if err := db.DeleteProduct(productID, userID, ctx); err != nil {
			if err.Error() == "Record Not found" {
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			// Real DB error
			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Product Deleted Successfully")

		resp := tools.Message {
			Message: "The product was deleted successfully",
			Data: true,
		}

		resp.WriteMessage(w)
	})

	router.Put("/{product}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.PUT./", ProductsRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("UserID", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		productID := chi.URLParam(r, "product")

		span.SetAttributes(
			attribute.String("ProductId", productID),
		)

		if strings.TrimSpace(productID) == "" {
			err := errors.New("Invalid product id")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		var product requests.CreateProduct
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := product.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		if err := db.UpdateProduct(productID, &product, userID, ctx); err != nil {
			if errors.Is(err, db.ErrProductNotFound){
				span.RecordError(db.ErrProductNotFound)
				span.SetStatus(codes.Error, db.ErrProductNotFound.Error())
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			// Real DB error
			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Updated Successfully")

		resp := tools.Message {
			Message: "The product was updated successfully",
			Data: true,
		}

		resp.WriteMessage(w)
	})

	router.Put("/{product}/image", func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s/create-account", AuthRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("UserID", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		productID := chi.URLParam(r, "product")

		span.SetAttributes(
			attribute.String("ProductId", productID),
		)
		
		if strings.TrimSpace(productID) == "" {
			err := errors.New("Invalid product id")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		_, err := db.GetProduct(userID, productID, ctx)
		if err != nil {
			if errors.Is(err, db.ErrProductNotFound){
				span.RecordError(db.ErrProductNotFound)
				span.SetStatus(codes.Error, db.ErrProductNotFound.Error())
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			log.Error("Failed to delete product: ", err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		// << 20 is a short of 2^20 -> 5 * 1MB
		r.ParseMultipartForm(5 << 20)

		file, handler, err := r.FormFile("image")
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("We couldn't get the image"))
			return
		}

		defer file.Close()
		
		contentType := handler.Header.Get("Content-Type")
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			err := errors.New("Only jpeg, png, and webp images are allowed")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		ext := filepath.Ext(handler.Filename)
		objectName := productID + ext

		path, putErr := platform.PutImage(objectName, file, handler.Size, contentType, ctx)
		if putErr != nil {
			span.RecordError(putErr)
			span.SetStatus(codes.Error, putErr.Error())
			msg := "We couldn't save your image"
			tools.InternalServerErrorHandler(w, &msg)
			return
		}

		if updateErr := db.UpdateProductImage(productID, path, userID, ctx); updateErr != nil {
			if errors.Is(updateErr, db.ErrProductNotFound){
				span.RecordError(db.ErrProductNotFound)
				span.SetStatus(codes.Error, db.ErrProductNotFound.Error())
				tools.NotFoundErrorHandler(w, "Product not found")
				return
			}

			span.RecordError(updateErr)
			span.SetStatus(codes.Error, putErr.Error())

			// Real DB error
			log.Error("Failed to update the product: ", updateErr)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Saved image")

		resp := tools.Message {
			Message: "The image was saved successfully",
			Data: path,
		}

		resp.WriteMessage(w)
	})
}


