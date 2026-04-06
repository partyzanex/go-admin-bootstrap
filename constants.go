package goadmin

const (
	DefaultAssetsPath    = "./assets"
	DefaultViewsPath     = "./views"
	DefaultLimit         = 20
	LoginRateLimitPerSec = 5
	SecureTokenLength    = 32

	UserContextKey   = "goadmin_user"
	DataContextKey   = "goadmin_data"
	LoggerContextKey = "goadmin_logger"

	AuthToken TokenType = "auth"

	UserNew     UserStatus = "new"
	UserActive  UserStatus = "active"
	UserBlocked UserStatus = "blocked"

	RoleOwner UserRole = "owner"
	RoleRoot  UserRole = "root"
	RoleUser  UserRole = "user"
)

const (
	DashboardURL = "/"
	LoginURL     = "/login"
	LogoutURL    = "/logout"

	UserListURL   = "/users"
	UserCreateURL = "/users/create"
	UserUpdateURL = "/users/:id/update"
	UserDeleteURL = "/users/:id/delete"

	DefaultAccessCookieName = "auth_token"
	DefaultMigrationsTable  = "goadmin_migrations"

	FaviconPrefix = "/favicon/:id"
)

const (
	adminPathVar   = "adminPath"
	loginURLVar    = "loginURL"
	logoutURLVar   = "logoutURL"
	userListURLVar = "userListURL"

	assetsRelativePath = "/assets"
)
