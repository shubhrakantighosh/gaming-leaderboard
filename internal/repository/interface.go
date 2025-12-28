package repository

import (
	"context"

	"gorm.io/gorm"

	"gaming-leaderboard/pkg/apperror"
)

type Interface[T any] interface {
	Create(
		ctx context.Context,
		entity *T,
	) (err apperror.Error)

	GetAll(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (results []*T, err apperror.Error)

	GetAllWithPagination(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (results []*T, count int64, err apperror.Error)

	Get(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (result T, err apperror.Error)
}
