package requests

import (
 "github.com/OscarVillanueva/goapi/internal/app/models/dao"
)

type ProductsResponse struct{
	Products []dao.Product `json:"products"`
	PageSize int `json:"page_size"`
	Pages int `json:"pages"`
}
