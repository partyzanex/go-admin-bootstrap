package goadmin

import "time"

const (
	DefaultAssetsPath    = "./assets"
	DefaultViewsPath     = "./views"
	DefaultLimit         = 20
	LoginRateLimitPerSec = 5
	SecureTokenLength    = 32

	// MinJWTSecretLen is the minimum byte length of JWTSecret for non-dev deployments.
	// HS256 provides its full 256-bit security only when the key is at least 32 bytes.
	MinJWTSecretLen = 32

	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 30 * 24 * time.Hour

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

	AuditLogURL = "/audit"

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
