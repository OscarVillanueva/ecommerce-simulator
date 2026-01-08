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


