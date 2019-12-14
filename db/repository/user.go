package repository

import "github.com/partyzanex/layer"

type userRepository struct {
	ex layer.BoilExecutor
}

func NewUserRepository(ex layer.BoilExecutor) *userRepository {
	return &userRepository{ex: ex}
}
