package postgres

import "github.com/volatiletech/sqlboiler/v4/queries/qm"

func mapSlice[T, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}

	return result
}

type filterOption func(mods []qm.QueryMod) []qm.QueryMod

func withWhereIn(column string, values []any) filterOption {
	return func(mods []qm.QueryMod) []qm.QueryMod {
		if len(values) > 0 {
			return append(mods, qm.WhereIn(column+" in ?", values...))
		}

		return mods
	}
}

func withWhereLike(column, value string) filterOption {
	return func(mods []qm.QueryMod) []qm.QueryMod {
		if value != "" {
			return append(mods, qm.Where(column+" like ?", "%"+value+"%"))
		}

		return mods
	}
}

func withWhereEq(column, value string) filterOption {
	return func(mods []qm.QueryMod) []qm.QueryMod {
		if value != "" {
			return append(mods, qm.Where(column+" = ?", value))
		}

		return mods
	}
}

func withLimit(limit int) filterOption {
	return func(mods []qm.QueryMod) []qm.QueryMod {
		if limit > 0 {
			return append(mods, qm.Limit(limit))
		}

		return mods
	}
}

func withOffset(offset int) filterOption {
	return func(mods []qm.QueryMod) []qm.QueryMod {
		if offset >= 0 {
			return append(mods, qm.Offset(offset))
		}

		return mods
	}
}

func buildQuery(base []qm.QueryMod, opts ...filterOption) []qm.QueryMod {
	mods := base
	for _, opt := range opts {
		mods = opt(mods)
	}

	return mods
}
