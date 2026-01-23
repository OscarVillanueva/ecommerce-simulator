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

func BatchPurchase(purchases *[]requests.CreatePurchase, buyer string, ctx context.Context,) (string, error) {
	db := platform.GetInstance()

	if db == nil {
		return "",errors.New("We couldn't connect to the database")
	}

	purchaseID := uuid.New().String()

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, purchase := range *purchases {
			err := tx.Transaction(func(subTx *gorm.DB) error {
				var product dao.Product
				if err := subTx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("uuid = ?", purchase.Product).First(&product).Error; err != nil {
					return err
				}

				if product.Quantity < purchase.Quantity {
					return &ErrInsufficientStock{Product: product.Name}
				}

				newPurchase := dao.Purchase{
					Uuid: purchaseID,
					Product: product.Uuid,
					Quantity: purchase.Quantity,
					Price: product.Price,
					PurchasedBy: buyer,
					CreatedAt: time.Now().UTC(),
				}

				if err := subTx.Create(&newPurchase).Error; err != nil {
					return err
				}

				quantity := product.Quantity - purchase.Quantity

				updatedProduct := map[string]interface{}{
					"quantity": quantity,
					"updated_at": time.Now().UTC(),
				}

				result := subTx.Model(&product).Where("uuid = ?", product.Uuid).Updates(updatedProduct)
				if result.Error != nil {
					return result.Error
				}

				if result.RowsAffected == 0 {
					return errors.New("Product not found")
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	return purchaseID, err
}
