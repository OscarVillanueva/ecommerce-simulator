package db

import (
	"fmt"
	"time"
	"context"
	"errors"

	"github.com/OscarVillanueva/goapi/internal/platform"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"

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

func BatchPurchase(purchases []requests.CreatePurchase, buyer string, ctx context.Context) (string, error) {
	db := platform.GetInstance()

	if db == nil {
		return "",errors.New("We couldn't connect to the database")
	}

	purchaseID := uuid.New().String()

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, purchase := range purchases {
			if purchase.Quantity < 0 {
				return errors.New("Quantity must be greater than zero")
			}

			var product dao.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("uuid = ?", purchase.Product).First(&product).Error; err != nil {
				return err
			}

			if product.Quantity < purchase.Quantity {
				return &ErrInsufficientStock{Product: product.Name}
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
				return err
			}

			quantity := product.Quantity - purchase.Quantity

			updatedProduct := map[string]interface{}{
				"quantity": quantity,
				"updated_at": time.Now().UTC(),
			}

			result := tx.Model(&product).Where("uuid = ?", product.Uuid).Updates(updatedProduct)
			if result.Error != nil {
				return result.Error
			}
		}

		return nil
	})

	return purchaseID, err
}

func FetchTickets(page int, ctx context.Context) (*[]dao.Ticket, error) {
	db := platform.GetInstance()

	if db == nil {
		return nil, errors.New("We couldn't connect to the database")
	}

	limit := 30
	offset := (page - 1) * limit

	var ticket []dao.Ticket
	err := db.Model(&dao.Purchase{}).
		Select("MAX(created_at) as created_at, SUM(price * quantity) as total, ticket_id").
		Group("ticket_id").
		Limit(limit).
		Offset(offset).
		Scan(&ticket).
		Error

	return &ticket, err
}
