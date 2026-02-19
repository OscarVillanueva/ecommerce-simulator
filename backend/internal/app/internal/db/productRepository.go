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
var RepositoryName = "product-repository"

func InsertProduct(product *requests.CreateProduct, belongTo string, ctx context.Context) (*dao.Product, error) {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.InsertProduct", RepositoryName))
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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	span.SetStatus(codes.Ok, "InsertProduct success")
	return &p, nil
}

func DeleteProduct(uuid string, user string, ctx context.Context) error  {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.DeleteProduct", RepositoryName))
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
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		span.RecordError(ErrProductNotFound)
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, "DeleteProduct successfully")

	return nil
}

func UpdateProduct(productID string, product *requests.CreateProduct, belongTo string, ctx context.Context) error  {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.DeleteProduct", RepositoryName))
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
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

func GetProducts(params parameters.GetProductsParams) (*requests.ProductsResponse, error) {
	tr := otel.Tracer(RepositoryName)
	ctx, span := tr.Start(params.Context, fmt.Sprintf("%s.GetProducts", RepositoryName))
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

	span.SetStatus(codes.Ok, "GetProducts successfully")

	return &response, err
}

func GetProduct(user string, productId string, ctx context.Context) (*dao.Product, error) {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.GetProducts", RepositoryName))
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
		span.RecordError(ErrProductNotFound)
		span.SetStatus(codes.Error, ErrProductNotFound.Error())
		return nil, ErrProductNotFound
	}

	span.SetStatus(codes.Ok, "GetProduct successfully")

	return &product, err
}

func UpdateProductImage(productID string, path string, userID string, ctx context.Context) error {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.GetProducts", RepositoryName))
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
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		span.RecordError(ErrProductNotFound)
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, "UpdateProductImage successfully")

	return nil
}

