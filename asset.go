package goadmin

type AssetKind uint8

const (
	JavaScript AssetKind = iota
	Stylesheet
	View
)

type Asset struct {
	Path      string
	SortOrder int
	Kind      AssetKind
}

var (
	JS = []*Asset{
		{"plugins/bootstrap/js/bootstrap.bundle.min.js", -800, JavaScript},
		{"js/admin.js", 0, JavaScript},
	}
	CSS = []*Asset{
		{"plugins/bootstrap/css/bootstrap.min.css", -1000, Stylesheet},
		{"css/style.css", -900, Stylesheet},
	}
	Views = []*Asset{
		{"layouts/nav.jet", 0, View},
		{"layouts/main.jet", 0, View},
		{"widgets/breadcrumbs.jet", 0, View},
		{"widgets/pagination.jet", 0, View},
		{"index/dashboard.jet", 0, View},
		{"errors/error.jet", 0, View},
		{"auth/login.jet", 0, View},
		{"user/form.jet", 0, View},
		{"user/index.jet", 0, View},
		{"audit/index.jet", 0, View},
	}
)
