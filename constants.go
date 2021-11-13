package goadmin

const (
	DefaultAssetsPath = "./assets"
	DefaultViewsPath  = "./views"
	DefaultLimit      = 20

	UserContextKey = "goadmin_user"
	DataContextKey = "goadmin_data"

	AuthToken TokenType = "auth"

	UserNew     UserStatus = "new"
	UserActive  UserStatus = "active"
	UserBlocked UserStatus = "blocked"

	RoleOwner UserRole = "owner"
	RoleRoot  UserRole = "root"
	RoleUser  UserRole = "user"
)

var (
	DashboardURL = "/"
	LoginURL     = "/login"
	LogoutURL    = "/logout"

	UserListURL   = "/users"
	UserCreateURL = "/users/create"
	UserUpdateURL = "/users/:id/update"
	UserDeleteURL = "/users/:id/delete"

	AccessCookieName = "auth_token"
	MigrationsTable  = "goadmin_migrations"
)
