package mongo

type DBName string
type Collection string

const (
	AuthDB   DBName = "auth_db"
	ConfigDB DBName = "config_db"
	CoreDB   DBName = "core_db"

	// Auth DB Collections
	AuditLogsCollection   Collection = "audit_logs"
	PermissionsCollection Collection = "permissions"
	RolesCollection       Collection = "roles"
	TenantsCollection     Collection = "tenants"
	UsersCollection       Collection = "users"

	// Config DB Collections
	ServiceConfigCollection Collection = "service_config"
	FeatureFlagsCollection  Collection = "feature_flags"
	EnvironmentCollection   Collection = "environment_settings"

	// Core DB Collections
	CategoriesCollection Collection = "categories"
	CustomerCollection   Collection = "customers"
	InventoryCollection  Collection = "inventory"
	OrderItemsCollection Collection = "order_items"
	OrdersCollection     Collection = "orders"
	ProductsCollection   Collection = "products"
	VendorsCollection    Collection = "vendors"
	WarehouseCollection  Collection = "warehouses"
)

var (
	dbToCollection = map[string][]string{
		string(AuthDB):   {string(AuditLogsCollection), string(PermissionsCollection), string(RolesCollection), string(TenantsCollection), string(UsersCollection)},
		string(ConfigDB): {string(ServiceConfigCollection), string(FeatureFlagsCollection), string(EnvironmentCollection)},
		string(CoreDB):   {string(CategoriesCollection), string(CustomerCollection), string(InventoryCollection), string(OrderItemsCollection), string(OrdersCollection), string(ProductsCollection), string(VendorsCollection), string(WarehouseCollection)},
	}
	collectionToDB = map[string]string{
		string(AuditLogsCollection):     string(AuthDB),
		string(PermissionsCollection):   string(AuthDB),
		string(RolesCollection):         string(AuthDB),
		string(TenantsCollection):       string(AuthDB),
		string(UsersCollection):         string(AuthDB),
		string(ServiceConfigCollection): string(ConfigDB),
		string(FeatureFlagsCollection):  string(ConfigDB),
		string(EnvironmentCollection):   string(ConfigDB),
		string(CategoriesCollection):    string(CoreDB),
		string(CustomerCollection):      string(CoreDB),
		string(InventoryCollection):     string(CoreDB),
		string(OrderItemsCollection):    string(CoreDB),
		string(OrdersCollection):        string(CoreDB),
		string(ProductsCollection):      string(CoreDB),
		string(VendorsCollection):       string(CoreDB),
		string(WarehouseCollection):     string(CoreDB),
	}
)
