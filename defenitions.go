package goadmin

import "github.com/pkg/errors"

const (
	DefaultAssetsPath = "./assets"
	DefaultViewsPath  = "./views"
	DefaultLimit      = 20

	Version = "v0.0.8"

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
	JS = []Asset{
		{
			"plugins/jquery/jquery-3.4.1.min.js",
			"https://code.jquery.com/jquery-3.4.1.min.js",
			-1000,
		},
		{
			"plugins/popper/popper.min.js",
			"https://cdn.jsdelivr.net/npm/popper.js@1.16.0/dist/umd/popper.min.js",
			-900,
		},
		{
			"plugins/bootstrap/js/bootstrap.min.js",
			"https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js",
			-800,
		},
	}
	CSS = []Asset{
		{
			"plugins/bootstrap/css/bootstrap.min.css",
			"https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css",
			-1000,
		},
		{
			"css/style.css",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/assets/css/style.css",
			-900,
		},
	}
	views = []Asset{
		{
			"layouts/nav.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/nav.jet",
			0,
		},

		{
			"layouts/main.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/main.jet",
			0,
		},
		{
			"widgets/breadcrumbs.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/breadcrumbs.jet",
			0,
		},
		{
			"widgets/pagination.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/pagination.jet",
			0,
		},
		{
			"index/dashboard.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/index/dashboard.jet",
			0,
		},
		{
			"errors/error.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/errors/error.jet",
			0,
		},
		{
			"auth/login.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/auth/login.jet",
			0,
		},
		{
			"user/form.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/user/form.jet",
			0,
		},
		{
			"user/index.jet",
			"https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/user/index.jet",
			0,
		},
	}
)
