package db

import (
	"time"
	"fmt"
	"math"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/models/parameters"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New("Product Not Found")
var ProductRepositoryName = "product-repository"

func InsertProduct(product *requests.CreateProduct, belongTo string, ctx context.Context) (*dao.Product, error) {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.InsertProduct", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return nil, dbErr
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

	if err := gorm.G[dao.Product](db).Create(trContext, &p); err != nil {
		span.SetAttributes(
			attribute.String("Uuid", p.Uuid),
			attribute.String("Name", product.Name),
			attribute.Int("Quantity", int(product.Quantity)),
			attribute.Float64("Price", float64(product.Price)),
			attribute.String("BelongsTo", belongTo),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	span.SetStatus(codes.Ok, fmt.Sprintf("%s.InsertProduct successfully", ProductRepositoryName))
	return &p, nil
}

func DeleteProduct(uuid string, user string, ctx context.Context) error  {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.DeleteProduct", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	result := db.WithContext(trContext).Where("uuid = (?) AND belongs_to = (?)", uuid, user).Delete(&dao.Product{})
	if result.Error != nil {
		span.SetAttributes(
			attribute.String("ProductUuid", uuid),
			attribute.String("BelongsTo", user),
		)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		span.RecordError(ErrProductNotFound)
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.DeleteProduct successfully", ProductRepositoryName))

	return nil
}

func UpdateProduct(productID string, product *requests.CreateProduct, belongTo string, ctx context.Context) error  {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.UpdateProduct", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	updatedProduct := map[string]interface{}{
		"name": product.Name,
		"price": product.Price,
		"quantity": product.Quantity,
		"updated_at": time.Now().UTC(),
	}

	result := db.Model(&dao.Product{}).
		WithContext(trContext).
		Where("uuid = (?) AND belongs_to = (?)", productID, belongTo).
		Updates(updatedProduct)
	
	if result.Error != nil {
		span.SetAttributes(
			attribute.String("ProductUuid", productID),
			attribute.String("Name", product.Name),
			attribute.Int("Quantity", int(product.Quantity)),
			attribute.Float64("Price", float64(product.Price)),
			attribute.String("BelongsTo", belongTo),
		)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.UpdateProduct successfully", ProductRepositoryName))

	return nil
}

func GetProducts(params parameters.GetProductsParams) (*requests.ProductsResponse, error) {
	tr := otel.Tracer(ProductRepositoryName)
	ctx, span := tr.Start(params.Context, fmt.Sprintf("%s.GetProducts", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return nil, dbErr
	}
	
	limit := 30
	offset := (params.Page - 1) * limit

	query := db.WithContext(ctx).Model(&dao.Product{}).Where("belongs_to = (?)", params.User)

	if params.OnlyAvailable {
		query = query.Where("quantity > 0")
		span.SetAttributes(
			attribute.Bool("OnlyAvailable", true),
		)
	}

	if params.SearchName != "" {
		fullSearch := "%" + params.SearchName + "%"
    query = query.Where("name LIKE ?", fullSearch)
		span.SetAttributes(
			attribute.String("SearchName", fullSearch),
		)
	}

	products := make([]dao.Product, 0)
	err := query.Limit(limit).Offset(offset).Find(&products).Error

	var count int
	countErr:= db.Model(&dao.Product{}).WithContext(params.Context).Select("COUNT(*)").Scan(&count).Error
	if countErr != nil {
		count = 1
		span.RecordError(countErr)
	}

	response := requests.ProductsResponse {
		Products: products,
		PageSize: limit,
		Pages: int(math.Ceil(float64(count) / float64(limit))),
	}

	return &response, err
}

func GetProduct(user string, productId string, ctx context.Context) (*dao.Product, error) {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.GetProduct", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()
	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return nil, dbErr
	}
	
	var product dao.Product
	err := db.WithContext(trContext).Where("belongs_to = (?) AND uuid = (?)", user, productId).First(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		span.SetAttributes(
			attribute.String("BelongsTo", user),
			attribute.String("ProductUuid", productId),
		)
		span.RecordError(ErrProductNotFound)
		span.SetStatus(codes.Error, ErrProductNotFound.Error())
		return nil, ErrProductNotFound
	}

	return &product, err
}

func UpdateProductImage(productID string, path string, userID string, ctx context.Context) error {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.UpdateProductImage", ProductRepositoryName))
	defer span.End()

	db := platform.GetInstance()
	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	updatedProduct := map[string]interface{}{
		"image": path,
	}

	result := db.Model(&dao.Product{}).
		WithContext(trContext).
		Where("uuid = (?) AND belongs_to = (?)", productID, userID).
		Updates(updatedProduct)
	
	if result.Error != nil {
		span.SetAttributes(
			attribute.String("BelongsTo", userID),
			attribute.String("ProductUuid", productID),
		)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		span.RecordError(ErrProductNotFound)
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.UpdateProductImage successfully", ProductRepositoryName))

	return nil
}

