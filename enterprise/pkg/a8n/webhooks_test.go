package a8n

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gh "github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Ran in integration_test.go
func testGitHubWebhook(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now()
		clock := func() time.Time { return now }

		ctx := context.Background()

		type state struct {
			ExternalServices []*repos.ExternalService
			Changesets       []*a8n.Changeset
			ChangesetEvents  []*a8n.ChangesetEvent
		}

		secret := "secret"
		s := state{
			ExternalServices: repos.ExternalServices{
				{
					Kind:        "GITHUB",
					DisplayName: "Github - With Webhook",
					Config: marshalJSON(t, &schema.GitHubConnection{
						Webhooks: []*schema.GitHubWebhook{
							{Org: "acme", Secret: secret},
						},
					}),
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			Changesets:      []*a8n.Changeset{},
			ChangesetEvents: []*a8n.ChangesetEvent{},
		}

		fs := loadFixtures(t)

		hook := &GitHubWebhook{
			Store: NewStoreWithClock(db, clock),
			Repos: repos.NewDBStore(db, sql.TxOptions{}),
			Now:   clock,
		}

		err := hook.Repos.UpsertExternalServices(ctx, s.ExternalServices...)
		if err != nil {
			t.Fatal(err)
		}

		err = hook.Store.CreateChangesets(ctx, s.Changesets...)
		if err != nil {
			t.Fatal(err)
		}

		err = hook.Store.UpsertChangesetEvents(ctx, s.ChangesetEvents...)
		if err != nil {
			t.Fatal(err)
		}

		for _, tc := range []struct {
			name   string
			secret string
			event  event
			code   int
			want   state
		}{
			{
				name:   "unauthorized",
				want:   s,
				secret: "wrong-secret",
				event:  fs["issue_comment-edited"],
				code:   http.StatusUnauthorized,
			},
			{
				name:   "non-existent-changeset",
				want:   s,
				secret: secret,
				event:  fs["issue_comment-edited"],
				code:   http.StatusOK,
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				body, err := json.Marshal(tc.event.event)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest("POST", "", bytes.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("X-Github-Event", tc.event.name)
				req.Header.Set("X-Hub-Signature", sign(body, []byte(tc.secret)))

				rec := httptest.NewRecorder()
				hook.ServeHTTP(rec, req)
				resp := rec.Result()

				if tc.code != 0 && tc.code != resp.StatusCode {
					bs, err := httputil.DumpResponse(resp, true)
					if err != nil {
						t.Fatal(err)
					}

					t.Log(string(bs))
					t.Errorf("have status code %d, want %d", resp.StatusCode, tc.code)
				}

				var have state
				have.ExternalServices, err = hook.Repos.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
				if err != nil {
					t.Fatal(err)
				}

				have.Changesets, _, err = hook.Store.ListChangesets(ctx, ListChangesetsOpts{Limit: 1000})
				if err != nil {
					t.Fatal(err)
				}

				have.ChangesetEvents, _, err = hook.Store.ListChangesetEvents(ctx, ListChangesetEventsOpts{Limit: 1000})
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

type event struct {
	name  string
	event interface{}
}

func loadFixtures(t testing.TB) map[string]event {
	t.Helper()

	matches, err := filepath.Glob("testdata/fixtures/*")
	if err != nil {
		t.Fatal(err)
	}

	fs := make(map[string]event, len(matches))
	for _, m := range matches {
		bs, err := ioutil.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}

		base := filepath.Base(m)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		parts := strings.SplitN(name, "-", 2)

		if len(parts) != 2 {
			t.Fatalf("unexpected fixture file name format: %s", m)
		}

		ev, err := gh.ParseWebHook(parts[0], bs)
		if err != nil {
			t.Fatal(err)
		}

		fs[name] = event{name: parts[0], event: ev}
	}

	return fs
}

func sign(message, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}
