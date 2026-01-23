package requests

type CreatePurchase struct {
	Product string `json:"product"`
	Quantity int32 `json:"quantity"`
}
