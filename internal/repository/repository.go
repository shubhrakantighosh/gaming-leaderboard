package repository

import (
	"context"
	"log"
	"net/http"

	"gaming-leaderboard/pkg/apperror"
	"gaming-leaderboard/pkg/db/postgres"
	"gaming-leaderboard/util"

	"gorm.io/gorm"
)

type Repository[T any] struct {
	Db *postgres.DbCluster
}

/*
Create
*/
func (r *Repository[T]) Create(
	ctx context.Context,
	entity *T,
) (err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.Create")

	tx := r.Db.GetMasterDB(ctx).Create(entity)
	if tx.Error != nil {
		log.Println(logTag, "Error while creating record:", tx.Error)
		return apperror.New(tx.Error, http.StatusBadRequest)
	}

	return apperror.Error{}
}

/*
Get All
*/
func (r *Repository[T]) GetAll(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []*T, err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.GetAll")

	tx := r.Db.GetSlaveDB(ctx).
		Model(new(T)).
		Where(filter).
		Scopes(scopes...).
		Find(&results)

	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)
		return nil, apperror.New(tx.Error, http.StatusBadRequest)
	}

	return results, apperror.Error{}
}

/*
Get All With Pagination
*/
func (r *Repository[T]) GetAllWithPagination(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []*T, count int64, err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.GetAllWithPagination")

	db := r.Db.GetSlaveDB(ctx).Model(new(T)).Where(filter).Scopes(scopes...)

	if errTx := db.Count(&count); errTx.Error != nil {
		log.Println(logTag, "Error while counting records:", errTx.Error)
		return nil, 0, apperror.New(errTx.Error, http.StatusBadRequest)
	}

	if count == 0 {
		return []*T{}, 0, apperror.Error{}
	}

	if errTx := db.Find(&results); errTx.Error != nil {
		log.Println(logTag, "Error while fetching records:", errTx.Error)
		return nil, 0, apperror.New(errTx.Error, http.StatusBadRequest)
	}

	return results, count, apperror.Error{}
}

/*
Get Single Record
*/
func (r *Repository[T]) Get(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (result T, err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.Get")

	tx := r.Db.GetSlaveDB(ctx).
		Model(new(T)).
		Where(filter).
		Scopes(scopes...).
		First(&result)

	if tx.Error != nil {
		log.Println(logTag, "Error while fetching record:", tx.Error)
		return result, apperror.New(tx.Error, http.StatusBadRequest)
	}

	return result, apperror.Error{}
}
