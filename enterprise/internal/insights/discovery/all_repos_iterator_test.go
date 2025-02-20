package discovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// TestAllReposIterator tests the AllReposIterator in the common use cases.
func TestAllReposIterator(t *testing.T) {
	ctx := context.Background()
	indexableReposLister := NewMockIndexableReposLister()
	repoStore := NewMockRepoStore()
	var timeOffset time.Duration
	clock := func() time.Time { return time.Now().Add(timeOffset) }

	// Mock the repo store listing, and confirm calls to it are cached.
	var (
		repoStoreListCalls []database.ReposListOptions
		nextRepoID         api.RepoID
	)
	repoStore.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]*types.Repo, error) {
		repoStoreListCalls = append(repoStoreListCalls, opt)
		var result []*types.Repo
		for i := 0; i < 3; i++ {
			nextRepoID++
			result = append(result, &types.Repo{ID: nextRepoID, Name: api.RepoName(fmt.Sprint(nextRepoID))})
		}
		if nextRepoID > 10 {
			return nil, nil
		}
		return result, nil
	})

	iter := NewAllReposIterator(indexableReposLister, repoStore, clock, false, 15*time.Minute, &prometheus.CounterOpts{Name: "fake_name123"})
	{
		// Do we get all 9 repositories?
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items0", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
	}

	trueP := true
	// Were the RepoStore.List calls as we expected?
	autogold.Want("repoStoreListCalls0", []database.ReposListOptions{
		{
			Index: &trueP,
			OrderBy: database.RepoListOrderBy{database.RepoListSort{
				Field: database.RepoListColumn("name"),
			}},
			LimitOffset: &database.LimitOffset{Limit: 1000},
		},
		{
			Index:   &trueP,
			OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
			LimitOffset: &database.LimitOffset{
				Limit:  1000,
				Offset: 3,
			},
		},
		{
			Index:   &trueP,
			OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
			LimitOffset: &database.LimitOffset{
				Limit:  1000,
				Offset: 6,
			},
		},
		{
			Index:   &trueP,
			OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
			LimitOffset: &database.LimitOffset{
				Limit:  1000,
				Offset: 9,
			},
		},
	}).Equal(t, repoStoreListCalls)

	// Again: do we get all 9 repositories, but this time all RepoStore.List calls were cached?
	repoStoreListCalls = nil
	nextRepoID = 0
	{
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items1", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
		autogold.Want("repoStoreListCalls1", []database.ReposListOptions{}).Equal(t, repoStoreListCalls)
	}

	// If the clock moves forward, does the cache expire and new RepoStore.List calls are made?
	timeOffset += iter.RepositoryListCacheTime
	repoStoreListCalls = nil
	nextRepoID = 0
	{
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items2", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
		autogold.Want("repoStoreListCalls2", []database.ReposListOptions{
			{
				Index: &trueP,
				OrderBy: database.RepoListOrderBy{database.RepoListSort{
					Field: database.RepoListColumn("name"),
				}},
				LimitOffset: &database.LimitOffset{Limit: 1000},
			},
			{
				Index:   &trueP,
				OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
				LimitOffset: &database.LimitOffset{
					Limit:  1000,
					Offset: 3,
				},
			},
			{
				Index:   &trueP,
				OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
				LimitOffset: &database.LimitOffset{
					Limit:  1000,
					Offset: 6,
				},
			},
			{
				Index:   &trueP,
				OrderBy: database.RepoListOrderBy{database.RepoListSort{Field: database.RepoListColumn("name")}},
				LimitOffset: &database.LimitOffset{
					Limit:  1000,
					Offset: 9,
				},
			},
		}).Equal(t, repoStoreListCalls)
	}
}

// TestAllReposIterator tests the AllReposIterator for Sourcegraph.com mode. Unfortunately, that
// mode is different because the repos list comes from a completely different table/store (this is
// technical debt for Sourcegraph.com, there is no reason the same store could not interface with
// both tables and handle pagination etc. the same way. The Search codebase also must deal with
// this cruft.)
func TestAllReposIterator_DotCom(t *testing.T) {
	ctx := context.Background()
	indexableReposLister := NewMockIndexableReposLister()
	repoStore := NewMockRepoStore()
	var timeOffset time.Duration
	clock := func() time.Time { return time.Now().Add(timeOffset) }

	// Mock the _default_ ("Sourcegraph.com") repo store listing, and confirm calls to it are cached.
	var (
		indexableReposListCall int // There is no pagination with this store! We'll probably want that, eventually.
		nextRepoID             api.RepoID
	)
	indexableReposLister.ListFunc.SetDefaultHook(func(ctx context.Context) ([]types.MinimalRepo, error) {
		indexableReposListCall++
		var result []types.MinimalRepo
		for i := 0; i < 9; i++ {
			nextRepoID++
			result = append(result, types.MinimalRepo{ID: nextRepoID, Name: api.RepoName(fmt.Sprint(nextRepoID))})
		}
		return result, nil
	})

	iter := NewAllReposIterator(indexableReposLister, repoStore, clock, true, 15*time.Minute, &prometheus.CounterOpts{Name: "fake_name456"})

	{
		// Do we get all 9 repositories?
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items0", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
	}

	// Were the IndexableRepos.List calls as we expected?
	autogold.Want("indexableReposStoreListCalls0", int(1)).Equal(t, indexableReposListCall)

	// Again: do we get all 9 repositories, but this time all IndexableRepos.List calls were cached?
	indexableReposListCall = 0
	nextRepoID = 0
	{
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items1", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
		autogold.Want("indexableReposStoreListCalls1", int(0)).Equal(t, indexableReposListCall)
	}

	// If the clock moves forward, does the cache expire and new IndexableRepos.List calls are made?
	timeOffset += iter.RepositoryListCacheTime
	indexableReposListCall = 0
	nextRepoID = 0
	{
		var each []string
		iter.ForEach(ctx, func(repoName string) error {
			each = append(each, repoName)
			return nil
		})
		autogold.Want("items2", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equal(t, each)
		autogold.Want("repoStoreListCalls2", int(1)).Equal(t, indexableReposListCall)
	}
}
