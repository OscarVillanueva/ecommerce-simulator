package handlers

import (
	"fmt"
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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/go-chi/chi/v5"
)

const PurchaseRouterName = "pruchase-router"

func PurchaseRouter(router chi.Router)  {
	router.Use(middleware.Authorization)

	router.Post("/", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.POST./", PurchaseRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("uuid", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		var ticket []requests.CreatePurchase
		if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		purchaseID, err := db.BatchPurchase(ticket, userID, ctx)

		if err != nil {
			var stockErr *db.ErrInsufficientStock
			if errors.As(err, &stockErr) {
				span.RecordError(errors.New(stockErr.Error()))
				span.SetStatus(codes.Error, stockErr.Error())
				tools.BadRequestErrorHandler(w, errors.New(stockErr.Error()))
				return
			}

			if err.Error() == "quantity must be greater than zero" {
				tools.BadRequestErrorHandler(w, err)
				return
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			tools.InternalServerErrorHandler(w, nil)
			return
		}


		span.SetStatus(codes.Ok, "Create Putchases successfully")

		resp := tools.Message {
			Message: "Purchase created successfully",
			Data: purchaseID,
		}

		resp.WriteMessage(w)
	})
	
	router.Get("/", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := 1

		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.GET./", ProductsRouterName))
		defer span.End()

		pageStr := r.URL.Query().Get("page")

		span.SetAttributes(
			attribute.String("Page", pageStr),
		)
		
		if pageStr != "" {
			parsedPage, err := strconv.Atoi(pageStr)

			if err != nil || parsedPage <= 0 {
				tools.BadRequestErrorHandler(w, errors.New("Invalid Page number"))
				return
			}

			page = parsedPage
		}
		
		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("UserUuid", userID),
		)

		if !ok || userID == ""{
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		tickets, err := db.FetchTickets(page, userID, ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Fetch Purchase successfully")

		resp := tools.Message {
			Message: "List of purchases",
			Data: tickets,
		}

		resp.WriteMessage(w)
	})

	router.Get("/{purchase}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.GET./", ProductsRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("uuid", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		purchaseID := chi.URLParam(r, "purchase")

		span.SetAttributes(
			attribute.String("PurchaseUuid", purchaseID),
		)

		if strings.TrimSpace(purchaseID) == "" {
			err := errors.New("Invalid purchase id")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		purchases, err := db.FetchPurchase(purchaseID, userID, ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if len(purchases) == 0 {
			span.SetStatus(codes.Unset, "Purchase not found")
			tools.NotFoundErrorHandler(w, "Purchase not found")
			return
		}

		span.SetStatus(codes.Ok, "Purchase found successfully")

		resp := tools.Message {
			Message: "Purchase details",
			Data: purchases,
		}

		resp.WriteMessage(w)
	})


	router.Delete("/{purchase}", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(ProductsRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s.DELETE./", ProductsRouterName))
		defer span.End()

		userID, ok := ctx.Value(middleware.UserUUIDKey).(string)

		span.SetAttributes(
			attribute.String("uuid", userID),
		)

		if !ok || userID == ""{
			err := errors.New("Missing user uuid")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		purchaseID := chi.URLParam(r, "purchase")

		span.SetAttributes(
			attribute.String("PurchaseUuid", purchaseID),
		)

		if strings.TrimSpace(purchaseID) == "" {
			err := errors.New("Invalid purchase id")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		purchases, err := db.FetchPurchase(purchaseID, userID, ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if len(purchases) == 0 {
			span.SetStatus(codes.Unset, err.Error())
			tools.NotFoundErrorHandler(w, "Purchase not found")
			return
		}

		if time.Since(purchases[0].CreatedAt) > time.Hour {
			err := errors.New("The purchase its to old to delete")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		if err := db.DeletePurchase(purchases, ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Purchase deleted successfully")

		resp := tools.Message {
			Message: "Purchase deleted successfully",
			Data: true,
		}

		resp.WriteMessage(w)
	})
}
