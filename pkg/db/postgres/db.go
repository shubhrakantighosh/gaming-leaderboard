package postgres

import (
	"context"
	"gaming-leaderboard/constants"
	"sync/atomic"

	"gorm.io/gorm"
)

type Db struct {
	*DbCluster
}

var dbInstance *Db

func GetCluster() *Db {
	return dbInstance
}

func SetCluster(cluster *DbCluster) {
	dbInstance = &Db{cluster}
}

type Consistency struct {
	consistency string
}

func (db *DbCluster) GetMasterDB(ctx context.Context) *gorm.DB {
	if val, ok := ctx.Value(constants.Consistency).(*Consistency); ok && val.consistency == constants.EventualConsistency {
		val.consistency = constants.StrongConsistency
	}

	return db.getMaster(ctx)
}

func (db *DbCluster) GetSlaveDB(ctx context.Context) *gorm.DB {
	if val, ok := ctx.Value(constants.Consistency).(*Consistency); ok && val.consistency == constants.StrongConsistency {
		return db.getMaster(ctx)
	}

	return db.getSlave(ctx)
}

func (db *DbCluster) getSlave(ctx context.Context) *gorm.DB {
	slavesCount := len(db.slaves)
	if slavesCount == 0 {
		return db.getMaster(ctx)
	}

	slaveNumber := int(atomic.AddUint64(&db.counter, 1) % uint64(slavesCount))
	return db.slaves[slaveNumber].db.WithContext(ctx)
}

func (db *DbCluster) getMaster(ctx context.Context) *gorm.DB {
	return db.master.db.WithContext(ctx)
}
