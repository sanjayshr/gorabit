package model

type Order struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	UserID   string `json:"user_id" bson:"user_id"`
	Product  string `json:"product" bson:"product"`
	Quantity int    `json:"quantity" bson:"quantity"`
}
