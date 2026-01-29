package db

import (
	"time"
	"math"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/models/parameters"
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

func GetProducts(params parameters.GetProductsParams) (*requests.ProductsResponse, error) {
	db := platform.GetInstance()

	if db == nil {
		return nil, errors.New("We couldn't connect to the database")
	}
	
	limit := 30
	offset := (params.Page - 1) * limit

	query := db.WithContext(params.Context).Model(&dao.Product{}).Where("belongs_to = (?)", params.User)

	if params.OnlyAvailable {
		query = query.Where("quantity > 0")
	}

	if params.SearchName != "" {
		fullSearch := "%" + params.SearchName + "%"
    query = query.Where("name LIKE ?", fullSearch)
	}

	products := make([]dao.Product, 0)
	err := query.Limit(limit).Offset(offset).Find(&products).Error

	var count int
	countErr:= db.Model(&dao.Product{}).WithContext(params.Context).Select("COUNT(*)").Scan(&count).Error
	if countErr != nil {
		count = 1
	}

	response := requests.ProductsResponse {
		Products: products,
		PageSize: limit,
		Pages: int(math.Ceil(float64(count) / float64(limit))),
	}

	return &response, err
}

func GetProduct(user string, productId string, ctx context.Context) (*dao.Product, error) {
	db := platform.GetInstance()

	if db == nil {
		return nil, errors.New("We couldn't connect to the database")
	}
	
	var product dao.Product
	err := db.WithContext(ctx).Where("belongs_to = (?) AND uuid = (?)", user, productId).First(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductNotFound
	}

	return &product, err
}

func UpdateProductImage(productID string, path string, userID string, ctx context.Context) error {
	db := platform.GetInstance()

	if db == nil {
		return errors.New("We couldn't connect to the database")
	}

	updatedProduct := map[string]interface{}{
		"image": path,
	}

	result := db.Model(&dao.Product{}).
		WithContext(ctx).
		Where("uuid = (?) AND belongs_to = (?)", productID, userID).
		Updates(updatedProduct)
	
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

