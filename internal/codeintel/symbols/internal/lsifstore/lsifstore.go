package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type LsifStore interface {
	// References
	GetReferenceLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error)

	// Implementation
	GetImplementationLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error)

	// Definition
	GetDefinitionLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error)

	// Monikers
	GetMonikersByPosition(ctx context.Context, uploadID int, path string, line, character int) (_ [][]precise.MonikerData, err error)
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, totalCount int, err error)

	// Packages
	GetPackageInformation(ctx context.Context, uploadID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error)
}

type store struct {
	db         *basestore.Store
	serializer *Serializer
	operations *operations
}

func New(db database.DB, observationContext *observation.Context) LsifStore {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		serializer: NewSerializer(),
		operations: newOperations(observationContext),
	}
}