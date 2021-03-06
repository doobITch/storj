// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package pointerdb

import (
	"context"

	"go.uber.org/zap"

	"storj.io/storj/internal/memory"
	"storj.io/storj/pkg/overlay"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/provider"
	"storj.io/storj/pkg/utils"
	"storj.io/storj/storage"
	"storj.io/storj/storage/boltdb"
	"storj.io/storj/storage/postgreskv"
	"storj.io/storj/storage/storelogger"
)

// CtxKeyPointerdb Used as pointerdb key
type CtxKeyPointerdb int

const (
	// BoltPointerBucket is the string representing the bucket used for `PointerEntries` in BoltDB
	BoltPointerBucket                 = "pointers"
	ctxKey            CtxKeyPointerdb = iota
	ctxKeyAllocation
)

// Config is a configuration struct that is everything you need to start a
// PointerDB responsibility
type Config struct {
	DatabaseURL          string      `help:"the database connection string to use" default:"bolt://$CONFDIR/pointerdb.db"`
	MinRemoteSegmentSize memory.Size `default:"1240" help:"minimum remote segment size"`
	MaxInlineSegmentSize memory.Size `default:"8000" help:"maximum inline segment size"`
	Overlay              bool        `default:"true" help:"toggle flag if overlay is enabled"`
	BwExpiration         int         `default:"45"   help:"lifespan of bandwidth agreements in days"`
}

// NewStore returns database for storing pointer data
func NewStore(dbURLString string) (db storage.KeyValueStore, err error) {
	driver, source, err := utils.SplitDBURL(dbURLString)
	if err != nil {
		return nil, err
	}
	if driver == "bolt" {
		db, err = boltdb.New(source, BoltPointerBucket)
	} else if driver == "postgresql" || driver == "postgres" {
		db, err = postgreskv.New(source)
	} else {
		err = Error.New("unsupported db scheme: %s", driver)
	}
	return db, err
}

// Run implements the provider.Responsibility interface
func (c Config) Run(ctx context.Context, server *provider.Provider) error {
	db, err := NewStore(c.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	cache := overlay.LoadFromContext(ctx)
	dblogged := storelogger.New(zap.L().Named("pdb"), db)

	service := NewService(zap.L(), dblogged)
	allocation := NewAllocationSigner(server.Identity(), c.BwExpiration)
	s := NewServer(zap.L(), service, allocation, cache, c, server.Identity())
	pb.RegisterPointerDBServer(server.GRPC(), s)
	// add the server to the context
	ctx = context.WithValue(ctx, ctxKey, service)
	ctx = context.WithValue(ctx, ctxKeyAllocation, allocation)
	return server.Run(ctx)
}

// LoadFromContext gives access to the pointerdb service from the context, or returns nil
func LoadFromContext(ctx context.Context) *Service {
	if v, ok := ctx.Value(ctxKey).(*Service); ok {
		return v
	}
	return nil
}

// LoadAllocationFromContext gives access to the payer bandwidth allocation service from the context, or returns nil
func LoadAllocationFromContext(ctx context.Context) *AllocationSigner {
	if v, ok := ctx.Value(ctxKeyAllocation).(*AllocationSigner); ok {
		return v
	}
	return nil
}
