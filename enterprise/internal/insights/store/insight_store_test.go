package store

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/hexops/valast"

	"github.com/inconshreveable/log15"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
)

func TestGet(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)

	_, err := timescale.Exec(`INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'test title', 'test description', 'unique-1'),
									       (2, 'test title 2', 'test description 2', 'unique-2')`)
	if err != nil {
		t.Fatal(err)
	}

	// assign some global grants just so the test can immediately fetch the created views
	_, err = timescale.Exec(`INSERT INTO insight_view_grants (insight_view_id, global)
									VALUES (1, true),
									       (2, true)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = timescale.Exec(`INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, last_snapshot_at, next_snapshot_after, deleted_at)
                            VALUES ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null),
									('series-id-3-deleted', 'query-3', $1, $1, $1, $1, $1, $1, $1);`, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = timescale.Exec(`INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
									VALUES (1, 1, 'label1', 'color1'),
											(1, 2, 'label2', 'color2'),
											(2, 2, 'second-label-2', 'second-color-2'),
											(2, 3, 'label3', 'color-2');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("test get all", func(t *testing.T) {
		store := NewInsightStore(timescale)

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-1",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-1",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label1",
				LineColor:           "color1",
				PresentationType:    types.Line,
			},
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-2",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-2",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label2",
				LineColor:           "color2",
				PresentationType:    types.Line,
			},
			{
				ViewID:              2,
				UniqueID:            "unique-2",
				SeriesID:            "series-id-2",
				Title:               "test title 2",
				Description:         "test description 2",
				Query:               "query-2",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "second-label-2",
				LineColor:           "second-color-2",
				PresentationType:    types.Line,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})

	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(timescale)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueIDs: []string{"unique-1"}})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-1",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-1",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label1",
				LineColor:           "color1",
				PresentationType:    types.Line,
			},
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-2",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-2",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label2",
				LineColor:           "color2",
				PresentationType:    types.Line,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(timescale)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: "unique-1"})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-1",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-1",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label1",
				LineColor:           "color1",
				PresentationType:    types.Line,
			},
			{
				ViewID:              1,
				UniqueID:            "unique-1",
				SeriesID:            "series-id-2",
				Title:               "test title",
				Description:         "test description",
				Query:               "query-2",
				CreatedAt:           now,
				OldestHistoricalAt:  now,
				LastRecordedAt:      now,
				NextRecordingAfter:  now,
				LastSnapshotAt:      now,
				NextSnapshotAfter:   now,
				SampleIntervalValue: 1,
				SampleIntervalUnit:  sampleIntervalUnit,
				Label:               "label2",
				LineColor:           "color2",
				PresentationType:    types.Line,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
}

func TestCreateSeries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2021, 5, 1, 1, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	ctx := context.Background()

	t.Run("test create series", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
		}

		got, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		want := types.InsightSeries{
			ID:                 1,
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			CreatedAt:          now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
		}

		log15.Info("values", "want", want, "got", got)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result from create insight series (want/got): %s", diff)
		}
	})
	t.Run("test create and get capture groups series", func(t *testing.T) {
		store := NewInsightStore(timescale)
		sampleIntervalUnit := "MONTH"
		_, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "capture-group-1",
			Query:                      "well hello there",
			Enabled:                    true,
			SampleIntervalUnit:         sampleIntervalUnit,
			SampleIntervalValue:        0,
			OldestHistoricalAt:         now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:             now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			CreatedAt:                  now,
			GeneratedFromCaptureGroups: true,
		})
		if err != nil {
			return
		}

		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			SeriesID: "capture-group-1",
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) < 1 {
			t.Fatal(err)
		}
		got[0].ID = 1 // normalizing this for test determinism

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestCreateView(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test create view", func(t *testing.T) {

		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}

		got, err := store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}

		want := types.InsightView{
			ID:               1,
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result from create insight view (want/got): %s", diff)
		}
	})
}

func TestCreateGetView_WithGrants(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CreateView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentationType: types.Line,
	}, []InsightViewGrant{UserGrant(1), OrgGrant(5)})
	if err != nil {
		t.Fatal(err)
	}
	series, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CreatedAt:          now,
		OldestHistoricalAt: now,
		LastRecordedAt:     now,
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		BackfillQueuedAt:   now,
		SampleIntervalUnit: string(types.Month),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "label1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("user 1 can see this view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserID: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for user 1 insight views")
		}
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("user 2 cannot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserID: []int{2}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for user 2 insight views")
		}
	})

	t.Run("org 1 cannot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserID: []int{3}, OrgID: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for org 1 insight views")
		}
	})
	t.Run("org 5 can see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserID: []int{3}, OrgID: []int{5}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for org 5 insight views")
		}
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
	t.Run("no users or orgs provided should only return global", func(t *testing.T) {
		uniqueID := "globalonly"
		view, err := store.CreateView(ctx, types.InsightView{
			Title:            "global only",
			Description:      "global only",
			UniqueID:         uniqueID,
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		series, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "globalseries",
			Query:              "global",
			CreatedAt:          now,
			OldestHistoricalAt: now,
			LastRecordedAt:     now,
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			BackfillQueuedAt:   now,
			SampleIntervalUnit: string(types.Month),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
			Label:  "label2",
			Stroke: "red",
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("unexpected count for global only insights")
		}
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestUpdateView(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test update view", func(t *testing.T) {
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		got, err := store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("AfterCreateView", types.InsightView{
			ID: 1, Title: "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}).Equal(t, got)

		include, exclude := "include repos", "exclude repos"
		got, err = store.UpdateView(ctx, types.InsightView{
			Title:    "new title",
			UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: &include,
				ExcludeRepoRegex: &exclude,
			},
			PresentationType: types.Line,
		})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("AfterUpdateView", types.InsightView{
			ID: 1, Title: "new title", UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: valast.Addr("include repos").(*string),
				ExcludeRepoRegex: valast.Addr("exclude repos").(*string),
			},
			PresentationType: types.PresentationType("LINE"),
		}).Equal(t, got)
	})
}

func TestUpdateViewSeries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test update view series", func(t *testing.T) {
		view, err := store.CreateView(ctx, types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		series, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
			Label:  "label",
			Stroke: "blue",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = store.UpdateViewSeries(ctx, series.SeriesID, view.ID, types.InsightViewSeriesMetadata{
			Label:  "new label",
			Stroke: "orange",
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("LabelAfterUpdateViewSeries", "new label").Equal(t, got[0].Label)
		autogold.Want("ColorAfterUpdateViewSeries", "orange").Equal(t, got[0].LineColor)
	})
}

func TestDeleteView(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CreateView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentationType: types.Line,
	}, []InsightViewGrant{GlobalGrant()})
	if err != nil {
		t.Fatal(err)
	}
	series, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CreatedAt:          now,
		OldestHistoricalAt: now,
		LastRecordedAt:     now,
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		BackfillQueuedAt:   now,
		SampleIntervalUnit: string(types.Month),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "label1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("delete view and check length", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) < 1 {
			t.Errorf("expected results before deleting view")
		}
		err = store.DeleteViewByUniqueID(ctx, uniqueID)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("expected results after deleting view")
		}
	})
}

func TestAttachSeriesView(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test attach and fetch", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricalAt:  now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		}
		series, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		view, err = store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		metadata := types.InsightViewSeriesMetadata{
			Label:  "my label",
			Stroke: "my stroke",
		}
		err = store.AttachSeriesToView(ctx, series, view, metadata)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{{
			ViewID:              1,
			UniqueID:            view.UniqueID,
			SeriesID:            series.SeriesID,
			Title:               view.Title,
			Description:         view.Description,
			Query:               series.Query,
			CreatedAt:           series.CreatedAt,
			OldestHistoricalAt:  series.OldestHistoricalAt,
			LastRecordedAt:      series.LastRecordedAt,
			NextRecordingAfter:  series.NextRecordingAfter,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalValue: 1,
			SampleIntervalUnit:  sampleIntervalUnit,
			Label:               "my label",
			LineColor:           "my stroke",
			PresentationType:    types.Line,
		}}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after attaching series to view (want/got): %s", diff)
		}
	})
}

func TestRemoveSeriesFromView(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test remove series from view", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricalAt:  now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		}
		series, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		view, err = store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		metadata := types.InsightViewSeriesMetadata{
			Label:  "my label",
			Stroke: "my stroke",
		}
		err = store.AttachSeriesToView(ctx, series, view, metadata)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{{
			ViewID:              1,
			UniqueID:            view.UniqueID,
			SeriesID:            series.SeriesID,
			Title:               view.Title,
			Description:         view.Description,
			Query:               series.Query,
			CreatedAt:           series.CreatedAt,
			OldestHistoricalAt:  series.OldestHistoricalAt,
			LastRecordedAt:      series.LastRecordedAt,
			NextRecordingAfter:  series.NextRecordingAfter,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalValue: 1,
			SampleIntervalUnit:  sampleIntervalUnit,
			Label:               "my label",
			LineColor:           "my stroke",
			PresentationType:    types.Line,
		}}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after attaching series to view (want/got): %s", diff)
		}

		err = store.RemoveSeriesFromView(ctx, series.SeriesID, view.ID)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		want = []types.InsightViewSeries{}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after removing series from view (want/got): %s", diff)
		}
		gotSeries, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: series.SeriesID, IncludeDeleted: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(gotSeries) == 0 || gotSeries[0].Enabled {
			t.Errorf("unexpected result: series does not exist or was not deleted after being removed from view")
		}
	})
}

func TestInsightStore_GetDataSeries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test empty", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected length of data series: %v", len(got))
		}
	})

	t.Run("test create and get series", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
		}
		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		want := []types.InsightSeries{created}

		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched insight data series want/got: %v", diff)
		}
	})
}

func TestInsightStore_StampRecording(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test create and update stamp", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
		}
		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		want := created
		want.LastRecordedAt = now
		want.NextRecordingAfter = time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)

		got, err := store.StampRecording(ctx, created)
		if err != nil {
			return
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched updated recording stamp want/got: %v", diff)
		}
	})
}

func TestInsightStore_StampBackfill(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	series := types.InsightSeries{
		SeriesID:           "unique-1",
		Query:              "query-1",
		OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
		LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		Enabled:            true,
		SampleIntervalUnit: string(types.Month),
	}
	created, err := store.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.StampBackfill(ctx, created)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("test only incomplete", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			BackfillIncomplete: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		want := 0
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched updated backfill_stamp count want/got: %v", diff)
		}
	})
	t.Run("test get all", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		want := 1
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched updated backfill_stamp count want/got: %v", diff)
		}
	})
}

func TestDirtyQueries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test read with no inserts", func(t *testing.T) {
		series := types.InsightSeries{
			ID:       1,
			SeriesID: "asdf",
			Query:    "qwerwre",
		}
		queries, err := store.GetDirtyQueries(ctx, &series)
		if err != nil {
			t.Fatal(err)
		}
		if len(queries) != 0 {
			t.Fatal("unexpected results of dirty queries")
		}
	})

	t.Run("write and read back", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:           "asdf",
			Query:              "qwerwre",
			SampleIntervalUnit: string(types.Month),
		}

		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		at := time.Date(2020, 1, 1, 5, 5, 5, 5, time.UTC).Truncate(time.Microsecond)

		if err := store.InsertDirtyQuery(ctx, &created, &types.DirtyQuery{
			ID:      1,
			Query:   created.Query,
			ForTime: at,
			Reason:  "this is a reason",
		}); err != nil {
			t.Fatal(err)
		}

		got, err := store.GetDirtyQueries(ctx, &created)
		if err != nil {
			t.Fatal(err)
		}
		want := []*types.DirtyQuery{
			{
				ID:      1,
				Query:   created.Query,
				ForTime: at,
				DirtyAt: now,
				Reason:  "this is a reason",
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched dirty query (want/got): %v", diff)
		}
	})
}

func TestDirtyQueriesAggregated(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test read with no inserts", func(t *testing.T) {
		series := types.InsightSeries{
			ID:       1,
			SeriesID: "asdf",
			Query:    "qwerwre",
		}
		queries, err := store.GetDirtyQueriesAggregated(ctx, series.SeriesID)
		if err != nil {
			t.Fatal(err)
		}
		if len(queries) != 0 {
			t.Fatal("unexpected results of dirty queries")
		}
	})

	t.Run("write and read back", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:           "asdf",
			Query:              "qwerwre",
			SampleIntervalUnit: string(types.Month),
		}

		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		at := time.Date(2020, 1, 1, 5, 5, 5, 5, time.UTC).Truncate(time.Microsecond)

		if err := store.InsertDirtyQuery(ctx, &created, &types.DirtyQuery{
			ID:      1,
			Query:   created.Query,
			ForTime: at,
			Reason:  "reason1",
		}); err != nil {
			t.Fatal(err)
		}
		if err := store.InsertDirtyQuery(ctx, &created, &types.DirtyQuery{
			ID:      1,
			Query:   created.Query,
			ForTime: at.AddDate(0, 0, 1),
			Reason:  "reason2",
		}); err != nil {
			t.Fatal(err)
		}
		if err := store.InsertDirtyQuery(ctx, &created, &types.DirtyQuery{
			ID:      1,
			Query:   created.Query,
			ForTime: at,
			Reason:  "reason1",
		}); err != nil {
			t.Fatal(err)
		}

		got, err := store.GetDirtyQueriesAggregated(ctx, created.SeriesID)
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestSetSeriesEnabled(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("start enabled set disabled set enabled", func(t *testing.T) {
		created, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "series1",
			Query:              "quer1",
			CreatedAt:          now,
			OldestHistoricalAt: now,
			LastRecordedAt:     now,
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			BackfillQueuedAt:   now,
			SampleIntervalUnit: string(types.Month),
		})
		if err != nil {
			t.Fatal(err)
		}
		if !created.Enabled {
			t.Errorf("series is disabled")
		}
		// set the series from enabled -> disabled
		err = store.SetSeriesEnabled(ctx, created.SeriesID, false)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{IncludeDeleted: true, SeriesID: created.SeriesID})
		if err != nil {
			t.Fatal()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching data series")
		}
		if got[0].Enabled {
			t.Errorf("series is enabled but should be disabled")
		}

		// set the series from disabled -> enabled
		err = store.SetSeriesEnabled(ctx, created.SeriesID, true)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDataSeries(ctx, GetDataSeriesArgs{IncludeDeleted: true, SeriesID: created.SeriesID})
		if err != nil {
			t.Fatal()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching data series")
		}
		if !got[0].Enabled {
			t.Errorf("series is enabled but should be disabled")
		}
	})
}

func TestFindMatchingSeries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CreatedAt:           now,
		OldestHistoricalAt:  now,
		LastRecordedAt:      now,
		NextRecordingAfter:  now,
		LastSnapshotAt:      now,
		NextSnapshotAfter:   now,
		BackfillQueuedAt:    now,
		SampleIntervalUnit:  string(types.Week),
		SampleIntervalValue: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("find a matching series when one exists", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 1", StepIntervalUnit: string(types.Week), StepIntervalValue: 1})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Equal(t, gotSeries, autogold.ExportedOnly())
		autogold.Want("FoundTrue", true).Equal(t, gotFound)
	})
	t.Run("find no matching series when none exist", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 2", StepIntervalUnit: string(types.Week), StepIntervalValue: 1})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Equal(t, gotSeries, autogold.ExportedOnly())
		autogold.Want("FoundFalse", false).Equal(t, gotFound)
	})
	t.Run("match capture group series", func(t *testing.T) {
		_, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "series id capture group",
			Query:                      "query 1",
			CreatedAt:                  now,
			OldestHistoricalAt:         now,
			LastRecordedAt:             now,
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			BackfillQueuedAt:           now,
			SampleIntervalUnit:         string(types.Week),
			SampleIntervalValue:        1,
			GeneratedFromCaptureGroups: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 1", StepIntervalUnit: string(types.Week), StepIntervalValue: 1, GenerateFromCaptureGroups: true})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Equal(t, gotSeries, autogold.ExportedOnly())
		autogold.Want("FoundTrueCaptureGroups", true).Equal(t, gotFound)
	})
}

func TestUpdateFrontendSeries(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CreatedAt:           now,
		OldestHistoricalAt:  now,
		LastRecordedAt:      now,
		NextRecordingAfter:  now,
		LastSnapshotAt:      now,
		NextSnapshotAfter:   now,
		BackfillQueuedAt:    now,
		SampleIntervalUnit:  string(types.Week),
		SampleIntervalValue: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("updates a series", func(t *testing.T) {
		gotBeforeUpdate, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("BeforeUpdateSeries", []types.InsightSeries{{
			ID:                  1,
			SeriesID:            "series id 1",
			Query:               "query 1",
			CreatedAt:           now,
			OldestHistoricalAt:  now,
			LastRecordedAt:      now,
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			Enabled:             true,
			SampleIntervalUnit:  "WEEK",
			SampleIntervalValue: 1,
		}}).Equal(t, gotBeforeUpdate)

		err = store.UpdateFrontendSeries(ctx, UpdateFrontendSeriesArgs{
			SeriesID:          "series id 1",
			Query:             "updated query!",
			StepIntervalUnit:  string(types.Month),
			StepIntervalValue: 5,
		})
		if err != nil {
			t.Fatal(err)
		}
		gotAfterUpdate, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("AfterUpdateSeries", []types.InsightSeries{{
			ID:                  1,
			SeriesID:            "series id 1",
			Query:               "updated query!",
			CreatedAt:           now,
			OldestHistoricalAt:  now,
			LastRecordedAt:      now,
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			Enabled:             true,
			SampleIntervalUnit:  "MONTH",
			SampleIntervalValue: 5,
		}}).Equal(t, gotAfterUpdate)
	})
}
