package parameters

import "context"

type GetProductsParams struct {
	User string
	Page int
	Context context.Context
	OnlyAvailable bool
	SearchName string
}

