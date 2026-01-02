package models

import "strings"

type Module string

const (
	ModuleAuth    Module = "Auth"
	ModuleConfig  Module = "Config"
	ModuleCore    Module = "Core"
	ModuleDB      Module = "DB"
	ModuleEvents  Module = "Events"
	ModuleGateway Module = "Gateway"
	ModuleSidecar Module = "Sidecar"
	ModuleWebUI   Module = "WebUI"
)

func IsValidModule(module string) bool {
	validModules := map[string]bool{
		strings.ToLower(string(ModuleAuth)):    true,
		strings.ToLower(string(ModuleConfig)):  true,
		strings.ToLower(string(ModuleCore)):    true,
		strings.ToLower(string(ModuleDB)):      true,
		strings.ToLower(string(ModuleEvents)):  true,
		strings.ToLower(string(ModuleGateway)): true,
		strings.ToLower(string(ModuleSidecar)): true,
		strings.ToLower(string(ModuleWebUI)):   true,
	}
	return validModules[strings.ToLower(module)]
}

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
