package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type DependenciesService interface {
	InsertPackageRepoRefs(ctx context.Context, deps []dependencies.MinimalPackageRepoRef) ([]dependencies.PackageRepoReference, []dependencies.PackageRepoRefVersion, error)
	ListPackageRepoFilters(ctx context.Context, opts dependencies.ListPackageRepoRefFiltersOpts) (_ []dependencies.PackageRepoFilter, hasMore bool, err error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type ExternalServiceStore interface {
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

type ReposStore interface {
	ListMinimalRepos(context.Context, database.ReposListOptions) ([]types.MinimalRepo, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, repoName api.RepoName, policies []policiesshared.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type PoliciesService interface {
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]policiesshared.ConfigurationPolicy, int, error)
}

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []uploadsshared.Index, err error)
	QueueIndexesForPackage(ctx context.Context, pkg dependencies.MinimialVersionedPackageRepo, assumeSynced bool) (err error)
}

type RepoUpdaterClient interface {
	RepoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
}

type AutoIndexingService interface {
	GetRepositoriesForIndexScan(ctx context.Context, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error)
}

type UploadService interface {
	GetUploadByID(ctx context.Context, id int) (shared.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) ([]uploadsshared.IndexesWithRepositoryNamespace, error)
}
