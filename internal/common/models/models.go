package models

// Address is a common address type used across core models
type Address struct {
	Street  string `bson:"street" json:"street"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Zip     string `bson:"zip" json:"zip"`
	Country string `bson:"country" json:"country"`
}

type SubscriptionLimits struct {
	MaxUsers          int `bson:"max_users" json:"max_users"`
	MaxProducts       int `bson:"max_products" json:"max_products"`
	MaxOrdersPerMonth int `bson:"max_orders_per_month" json:"max_orders_per_month"`
	StorageGB         int `bson:"storage_gb" json:"storage_gb"`
}
