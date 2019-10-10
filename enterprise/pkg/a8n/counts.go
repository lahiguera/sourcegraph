package a8n

import (
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

type ChangesetCounts struct {
	Time                 time.Time
	Total                int32
	Merged               int32
	Closed               int32
	Open                 int32
	OpenApproved         int32
	OpenChangesRequested int32
	OpenPending          int32
}

func (cc *ChangesetCounts) String() string {
	return fmt.Sprintf("%s (Total: %d, Merged: %d, Closed: %d, Open: %d, OpenApproved: %d, OpenChangesRequested: %d, OpenPending: %d)",
		cc.Time.String(),
		cc.Total,
		cc.Merged,
		cc.Closed,
		cc.Open,
		cc.OpenApproved,
		cc.OpenChangesRequested,
		cc.OpenPending,
	)
}

type Event interface {
	Timestamp() time.Time
	Type() a8n.ChangesetEventKind
	Changeset() int64
}

type Events []Event

func (es Events) Len() int      { return len(es) }
func (es Events) Swap(i, j int) { es[i], es[j] = es[j], es[i] }

// Less sorts events by their timestamps
func (es Events) Less(i, j int) bool {
	return es[i].Timestamp().Before(es[j].Timestamp())
}

func CalcCounts(start, end time.Time, cs []*a8n.Changeset, es ...Event) ([]*ChangesetCounts, error) {
	ts := generateTimestamps(start, end)
	counts := make([]*ChangesetCounts, len(ts))
	for i, t := range ts {
		counts[i] = &ChangesetCounts{Time: t}
	}

	// Sort all events once by their timestamps
	events := Events(es)
	sort.Sort(events)

	// Map sorted events to their changesets
	byChangeset := make(map[*a8n.Changeset]Events)
	for _, c := range cs {
		group := Events{}
		for _, e := range events {
			if e.Changeset() == c.ID {
				group = append(group, e)
			}
		}
		byChangeset[c] = group
	}

	for c, csEvents := range byChangeset {
		// We don't have an event for "open", so we check when it was
		// created on codehost
		openedAt, err := c.ExternalCreatedAt()
		if err != nil {
			return nil, err
		}

		// For each changeset and its events, go through every point in time we
		// want to record and reconstruct the changesets history until that
		// point
		for _, count := range counts {
			t := count.Time

			if openedAt.Before(t) || openedAt.Equal(t) {
				count.Total++
				count.Open++
			} else {
				// No need to look at events if changeset was not created yet
				continue
			}

			for _, e := range csEvents {
				// Event happened after point in time we're looking at, ignore
				if e.Timestamp().After(t) {
					continue
				}
				switch e.Type() {
				case a8n.ChangesetEventKindGitHubClosed:
					count.Open--
					count.Closed++
				case a8n.ChangesetEventKindGitHubReopened:
					count.Open++
					count.Closed--
				case a8n.ChangesetEventKindGitHubMerged:
					count.Merged++
					count.Open--
				}
			}
		}
	}

	return counts, nil
}

func generateTimestamps(start, end time.Time) []time.Time {
	// Walk backwards from `end` to >= `start` in 1 day intervals
	// Backwards so we always end exactly on `end`
	ts := []time.Time{}
	for t := end; t.After(start) || t.Equal(start); t = t.Add(-24 * time.Hour) {
		ts = append(ts, t)
	}

	// Now reverse so we go from oldest to newest in slice
	for i := len(ts)/2 - 1; i >= 0; i-- {
		opp := len(ts) - 1 - i
		ts[i], ts[opp] = ts[opp], ts[i]
	}

	return ts
}
