package db

import (
	"errors"
	"time"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func InsertProduct(product *requests.CreateProduct, belongTo string, ctx context.Context) (*dao.Product, error) {
	db := platform.GetInstance()

	if db == nil {
		return nil, errors.New("We couldn't connect to the database")
	}

	p := dao.Product{
		Uuid: uuid.New().String(),
		Name: product.Name,
		Price: product.Price,
		Quantity: product.Quantity,
		Image: nil,
		BelongsTo: belongTo,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: nil,
	}

	if err := gorm.G[dao.Product](db).Create(ctx, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func DeleteProduct(uuid string, user string, ctx context.Context) error  {
	db := platform.GetInstance()

	if db == nil {
		return errors.New("We couldn't connect to the database")
	}

	result := db.WithContext(ctx).Where("uuid = (?) AND belongs_to = (?)", uuid, user).Delete(&dao.Product{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("Record Not found")
	}

	return nil
}

func UpdateProduct(productID string, product *requests.CreateProduct, belongTo string, ctx context.Context) error  {
	db := platform.GetInstance()

	if db == nil {
		return errors.New("We couldn't connect to the database")
	}

	updatedProduct := dao.Product{
		Name: product.Name,
		Price: product.Price,
		Quantity: product.Quantity,
		BelongsTo: belongTo,
	}

	result := db.Model(&updatedProduct).
		WithContext(ctx).
		Where("uuid = (?) AND belongs_to = (?)", productID, belongTo).
		Updates(updatedProduct)
	
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("Record Not found")
	}

	return nil
}

