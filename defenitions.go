package goadmin

import "github.com/pkg/errors"

const (
	DefaultAssetsPath = "./assets"
	DefaultViewsPath  = "./views"
	DefaultLimit      = 20

	Version = "v0.0.2"

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

var (
	ErrInvalidPort          = errors.New("invalid http port")
	ErrRequiredDB           = errors.New("required database connection instance")
	ErrContextNotConfigured = errors.New("admin context not configured")
	ErrRequiredUserName     = errors.New("required user name")
	ErrRequiredUserLogin    = errors.New("required user login")
	ErrInvalidUserLogin     = errors.New("invalid user login")
	ErrInvalidUserStatus    = errors.New("invalid user status")
	ErrRequiredUserID       = errors.New("required user id")
	ErrRequiredUserPassword = errors.New("required user password")
	ErrWrongPassword        = errors.New("wrong password")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTokenExpired         = errors.New("token expired")
)

var (
	js = []Asset{
		{
			"plugins/jquery/jquery-3.4.1.min.js",
			"https://code.jquery.com/jquery-3.4.1.min.js",
		},
		{
			"plugins/bootstrap/js/bootstrap.min.js",
			"https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js",
		},
	}
	css = []Asset{
		{
			"plugins/bootstrap/css/bootstrap.min.css",
			"https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css",
		},
		{
			"css/style.css",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/assets/css/style.css",
		},
	}
	views = []Asset{
		{
			"layouts/nav.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/nav.jet"},
		{
			"layouts/main.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/main.jet",
		},
		{
			"widgets/breadcrumbs.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/breadcrumbs.jet",
		},
		{
			"widgets/pagination.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/pagination.jet",
		},
		{
			"index/dashboard.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/index/dashboard.jet",
		},
		{
			"errors/error.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/errors/error.jet",
		},
		{
			"auth/login.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/auth/login.jet",
		},
		{
			"user/form.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/user/form.jet",
		},
		{
			"user/index.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/user/index.jet",
		},
	}
)
