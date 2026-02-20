package db

import (
	"fmt"
	"time"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/platform"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
	"gorm.io/gorm"
)

type ErrInsufficientStock struct {
    Product string
}

func (e *ErrInsufficientStock) Error() string {
    return fmt.Sprintf("insufficient stock for %s", e.Product)
}

const PurchaseRepositoryName = "purchase-repository"

func BatchPurchase(purchases []requests.CreatePurchase, buyer string, ctx context.Context) (string, error) {
	tr := otel.Tracer(PurchaseRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.BatchPurchase", PurchaseRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return "", dbErr
	}

	purchaseID := uuid.New().String()

	err := db.WithContext(trContext).Transaction(func(tx *gorm.DB) error {
		for _, purchase := range purchases {
			if purchase.Quantity < 0 {
				quantityErr := errors.New("Quantity must be greater than zero: " + purchase.Product)
				span.RecordError(quantityErr)
				span.SetStatus(codes.Error, quantityErr.Error())
				return quantityErr
			}

			var product dao.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("uuid = ?", purchase.Product).First(&product).Error; err != nil {
				span.SetAttributes(
					attribute.String("ProductUuid", purchase.Product),
				)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}

			if product.Quantity < purchase.Quantity {
				msgErr := ErrInsufficientStock{Product: product.Name}
				span.SetAttributes(
					attribute.String("ProductUuid", product.Uuid),
					attribute.Int("PurchaseQuantity", int(purchase.Quantity)),
				)
				span.RecordError(errors.New(msgErr.Error()))
				span.SetStatus(codes.Error, msgErr.Error())
				return &msgErr
			}

			newPurchase := dao.Purchase{
				Uuid: uuid.New().String(),
				TicketId: purchaseID,
				Product: product.Uuid,
				Quantity: purchase.Quantity,
				Price: product.Price,
				PurchasedBy: buyer,
				CreatedAt: time.Now().UTC(),
			}

			if err := tx.Create(&newPurchase).Error; err != nil {
				span.SetAttributes(
					attribute.String("Uuid", newPurchase.Uuid),
					attribute.String("TicketId", newPurchase.TicketId),
					attribute.String("ProductUuid", newPurchase.Product),
					attribute.Float64("Price", float64(newPurchase.Price)),
					attribute.Int("Quantity", int(newPurchase.Quantity)),
					attribute.String("PurchasedBy", newPurchase.PurchasedBy),
				)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}

			quantity := product.Quantity - purchase.Quantity

			updatedProduct := map[string]interface{}{
				"quantity": quantity,
				"updated_at": time.Now().UTC(),
			}

			result := tx.Model(&product).Where("uuid = ?", product.Uuid).Updates(updatedProduct)
			if result.Error != nil {
				span.SetAttributes(
					attribute.String("ProductUuid", product.Uuid),
					attribute.Int("Quantity", int(quantity)),
				)
				span.RecordError(result.Error)
				span.SetStatus(codes.Error, fmt.Sprintf("Unable to update the quantity of product: %s", product.Uuid))
				return result.Error
			}
		}

		return nil
	})

	return purchaseID, err
}

func FetchTickets(page int, buyer string, ctx context.Context) ([]dao.Ticket, error) {
	tr := otel.Tracer(PurchaseRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.FetchTickets", PurchaseRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return nil, dbErr
	}

	limit := 30
	offset := (page - 1) * limit

	span.SetAttributes(
		attribute.String("PurchasedBy", buyer),
	)

	ticket := make([]dao.Ticket, 0)
	err := db.Model(&dao.Purchase{}).
		WithContext(trContext).
		Where("purchased_by = ?", buyer).
		Select("MAX(created_at) as created_at, SUM(price * quantity) as total, ticket_id").
		Group("ticket_id").
		Limit(limit).
		Offset(offset).
		Scan(&ticket).
		Error

	return ticket, err
}

func FetchPurchase(purchaseId string, buyer string, ctx context.Context) ([]dao.Purchase, error) {
	tr := otel.Tracer(PurchaseRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.FetchPurchase", PurchaseRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return nil, dbErr
	}

	span.SetAttributes(
		attribute.String("TicketId", purchaseId),
		attribute.String("PurchasedBy", buyer),
	)

	purchases := make([]dao.Purchase, 0)
	err := db.Model(&dao.Purchase{}).
		WithContext(trContext).
		Where("ticket_id = ? AND purchased_by = ?", purchaseId, buyer).
		Find(&purchases).
		Error

	return purchases, err
}

func DeletePurchase(purchases []dao.Purchase, ctx context.Context) error  {
	tr := otel.Tracer(PurchaseRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.DeletePurchase", PurchaseRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	err := db.WithContext(trContext).Transaction(func(tx *gorm.DB) error {
		for _, purchase := range purchases {
			var product dao.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("uuid = ?", purchase.Product).First(&product).Error; err != nil {
				span.SetAttributes(
					attribute.String("ProductUuid", purchase.Product),
				)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}

			quantity := product.Quantity + purchase.Quantity

			updatedProduct := map[string]interface{}{
				"quantity": quantity,
				"updated_at": time.Now().UTC(),
			}

			result := tx.Model(&product).Where("uuid = ?", product.Uuid).Updates(updatedProduct)
			if result.Error != nil {
				span.SetAttributes(
					attribute.String("ProductUuid", product.Uuid),
				)
				span.RecordError(result.Error)
				span.SetStatus(codes.Error, fmt.Sprintf("We couldn't update the quantity of the product: %s", purchase.Product))
				return result.Error
			}

			err := tx.Where("uuid = ? AND ticket_id = ?", purchase.Uuid, purchase.TicketId).Delete(&purchase).Error
			if err != nil {
				span.SetAttributes(
					attribute.String("PurchaseUuid", purchase.Uuid),
					attribute.String("TicketId", purchase.TicketId),
				)
				span.RecordError(err)
				span.SetStatus(codes.Error, fmt.Sprintf("We couldn't the purchase: %s", purchase.Uuid))
				return err
			}
		}

		return nil
	})

	return err
}
