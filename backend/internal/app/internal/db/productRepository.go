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

var ErrProductNotFound = errors.New("Product Not Found")

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
		return ErrProductNotFound
	}

	return nil
}

func UpdateProduct(productID string, product *requests.CreateProduct, belongTo string, ctx context.Context) error  {
	db := platform.GetInstance()

	if db == nil {
		return errors.New("We couldn't connect to the database")
	}

	updatedProduct := map[string]interface{}{
		"name": product.Name,
		"price": product.Price,
		"quantity": product.Quantity,
		"updated_at": time.Now().UTC(),
	}

	result := db.Model(&dao.Product{}).
		WithContext(ctx).
		Where("uuid = (?) AND belongs_to = (?)", productID, belongTo).
		Updates(updatedProduct)
	
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

func GetProducts(user string, ctx context.Context) (*[]dao.Product, error) {
	db := platform.GetInstance()

	if db == nil {
		return nil, errors.New("We couldn't connect to the database")
	}
	
	var products []dao.Product
	if err := db.WithContext(ctx).Where("belongs_to = (?)", user).Find(&products).Error; err != nil {
		return nil, err
	}

	return &products, nil
}
