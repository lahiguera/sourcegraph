package a8n

import (
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

var update = flag.Bool("update", false, "update testdata")

func TestCalcCounts(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	ghChangesetCreated := func(id int64, t time.Time) *a8n.Changeset {
		return &a8n.Changeset{ID: id, Metadata: &github.PullRequest{CreatedAt: t}}
	}

	tests := []struct {
		name       string
		changesets []*a8n.Changeset
		start      time.Time
		end        time.Time
		events     []Event
		want       []*ChangesetCounts
	}{
		{
			name: "single changeset open merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "multiple changesets open merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(2)),
				ghChangesetCreated(2, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 2, Open: 2},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			name: "multiple changesets open merged different times",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(3)),
				ghChangesetCreated(2, daysAgo(2)),
			},
			start: daysAgo(4),
			events: []Event{
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1},
				{Time: daysAgo(2), Total: 2, Open: 1, Merged: 1},
				{Time: daysAgo(1), Total: 2, Merged: 2},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			name: "changeset merged and closed at same time",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "changeset merged and closed at same time, reversed order in slice",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(2)),
			},
			start: daysAgo(2),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(2), Total: 1, Open: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "single changeset open closed reopened merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "multiple changesets open closed reopened merged different times",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(5)),
				ghChangesetCreated(2, daysAgo(4)),
			},
			start: daysAgo(6),
			events: []Event{
				fakeEvent{t: daysAgo(4), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 2},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 2},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(0), kind: a8n.ChangesetEventKindGitHubMerged, id: 2},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(6), Total: 0, Open: 0},
				{Time: daysAgo(5), Total: 1, Open: 1},
				{Time: daysAgo(4), Total: 2, Open: 1, Closed: 1},
				{Time: daysAgo(3), Total: 2, Open: 1, Closed: 1},
				{Time: daysAgo(2), Total: 2, Open: 2},
				{Time: daysAgo(1), Total: 2, Open: 1, Merged: 1},
				{Time: daysAgo(0), Total: 2, Merged: 2},
			},
		},
		{
			name: "single changeset open closed reopened merged, unsorted events",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(4)),
			},
			start: daysAgo(5),
			events: []Event{
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
				fakeEvent{t: daysAgo(3), kind: a8n.ChangesetEventKindGitHubClosed, id: 1},
				fakeEvent{t: daysAgo(2), kind: a8n.ChangesetEventKindGitHubReopened, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(5), Total: 0, Open: 0},
				{Time: daysAgo(4), Total: 1, Open: 1},
				{Time: daysAgo(3), Total: 1, Open: 0, Closed: 1},
				{Time: daysAgo(2), Total: 1, Open: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "single changeset open, pending review, approved, merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				&a8n.ChangesetEvent{
					ChangesetID: 1,
					Kind:        a8n.ChangesetEventKindGitHubReviewed,
					Metadata: &github.PullRequestReview{
						UpdatedAt: daysAgo(2),
						State:     "APPROVED",
					},
				},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenApproved: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "single changeset open, pending review, changes requested, merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				&a8n.ChangesetEvent{
					ChangesetID: 1,
					Kind:        a8n.ChangesetEventKindGitHubReviewed,
					Metadata: &github.PullRequestReview{
						UpdatedAt: daysAgo(2),
						State:     "CHANGES_REQUESTED",
					},
				},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenChangesRequested: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
		{
			name: "single changeset open, pending review, pending, merged",
			changesets: []*a8n.Changeset{
				ghChangesetCreated(1, daysAgo(3)),
			},
			start: daysAgo(4),
			events: []Event{
				&a8n.ChangesetEvent{
					ChangesetID: 1,
					Kind:        a8n.ChangesetEventKindGitHubReviewed,
					Metadata: &github.PullRequestReview{
						UpdatedAt: daysAgo(2),
						State:     "PENDING",
					},
				},
				fakeEvent{t: daysAgo(1), kind: a8n.ChangesetEventKindGitHubMerged, id: 1},
			},
			want: []*ChangesetCounts{
				{Time: daysAgo(4), Total: 0, Open: 0},
				{Time: daysAgo(3), Total: 1, Open: 1},
				{Time: daysAgo(2), Total: 1, Open: 1, OpenPending: 1},
				{Time: daysAgo(1), Total: 1, Merged: 1},
				{Time: daysAgo(0), Total: 1, Merged: 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.end.IsZero() {
				tc.end = now
			}

			have, err := CalcCounts(tc.start, tc.end, tc.changesets, tc.events...)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf("wrong counts calculated. diff=%s", cmp.Diff(have, tc.want))
			}
		})
	}
}

func parse(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
}

type fakeEvent struct {
	t    time.Time
	kind a8n.ChangesetEventKind
	id   int64
}

func (e fakeEvent) Timestamp() time.Time         { return e.t }
func (e fakeEvent) Type() a8n.ChangesetEventKind { return e.kind }
func (e fakeEvent) Changeset() int64             { return e.id }
