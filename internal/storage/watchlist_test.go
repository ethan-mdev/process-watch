package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethan-mdev/service-watch/internal/core"
)

func TestNewJSONWatchlist_LoadMissingFileStartsEmpty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	items, err := wl.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(items))
	}
}

func TestNewJSONWatchlist_LoadInvalidJSONDoesNotPanicAndStartsEmpty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "watchlist.json")
	if err := os.WriteFile(path, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to seed invalid json: %v", err)
	}

	wl := NewJSONWatchlist(path)
	items, err := wl.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(items))
	}
}

func TestAddGetListRemoveCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	entryA := testEntry("svc-a")
	entryB := testEntry("svc-b")

	if err := wl.Add(ctx, entryA); err != nil {
		t.Fatalf("Add(entryA) returned error: %v", err)
	}
	if err := wl.Add(ctx, entryB); err != nil {
		t.Fatalf("Add(entryB) returned error: %v", err)
	}

	gotA, err := wl.Get(ctx, entryA.Name)
	if err != nil {
		t.Fatalf("Get(%q) returned error: %v", entryA.Name, err)
	}
	if gotA.Name != entryA.Name || gotA.RestartCmd != entryA.RestartCmd {
		t.Fatalf("Get(%q) mismatch: got %+v", entryA.Name, gotA)
	}

	items, err := wl.List(ctx)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(items))
	}

	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Name] = true
	}
	if !seen[entryA.Name] || !seen[entryB.Name] {
		t.Fatalf("List() missing expected entries: %+v", seen)
	}

	if err := wl.Remove(ctx, entryA.Name); err != nil {
		t.Fatalf("Remove(%q) returned error: %v", entryA.Name, err)
	}
	if _, err := wl.Get(ctx, entryA.Name); err == nil {
		t.Fatalf("Get(%q) expected error after removal", entryA.Name)
	}
}

func TestAddDuplicateAndMissingEntityErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	entry := testEntry("svc-a")
	if err := wl.Add(ctx, entry); err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	if err := wl.Add(ctx, entry); err == nil {
		t.Fatalf("Add(duplicate) expected error")
	}
	if err := wl.Remove(ctx, "missing"); err == nil {
		t.Fatalf("Remove(missing) expected error")
	}
	if _, err := wl.Get(ctx, "missing"); err == nil {
		t.Fatalf("Get(missing) expected error")
	}
	if err := wl.Update(ctx, "missing", false); err == nil {
		t.Fatalf("Update(missing) expected error")
	}
	if err := wl.IncrementFailCount(ctx, "missing"); err == nil {
		t.Fatalf("IncrementFailCount(missing) expected error")
	}
	if err := wl.IncrementRestartCount(ctx, "missing"); err == nil {
		t.Fatalf("IncrementRestartCount(missing) expected error")
	}
	if err := wl.ResetFailCount(ctx, "missing"); err == nil {
		t.Fatalf("ResetFailCount(missing) expected error")
	}
}

func TestUpdateAndCounterMutationsPersist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	entry := testEntry("svc-counters")
	if err := wl.Add(ctx, entry); err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	if err := wl.Update(ctx, entry.Name, false); err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}
	if err := wl.IncrementFailCount(ctx, entry.Name); err != nil {
		t.Fatalf("IncrementFailCount() returned error: %v", err)
	}
	if err := wl.IncrementFailCount(ctx, entry.Name); err != nil {
		t.Fatalf("IncrementFailCount() returned error: %v", err)
	}

	beforeRestart := time.Now()
	if err := wl.IncrementRestartCount(ctx, entry.Name); err != nil {
		t.Fatalf("IncrementRestartCount() returned error: %v", err)
	}
	if err := wl.ResetFailCount(ctx, entry.Name); err != nil {
		t.Fatalf("ResetFailCount() returned error: %v", err)
	}

	got, err := wl.Get(ctx, entry.Name)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if got.AutoRestart {
		t.Fatalf("AutoRestart = true, want false")
	}
	if got.RestartCount != 1 {
		t.Fatalf("RestartCount = %d, want 1", got.RestartCount)
	}
	if got.FailCount != 0 {
		t.Fatalf("FailCount = %d, want 0", got.FailCount)
	}

	ts, err := time.Parse(time.RFC3339, got.LastRestart)
	if err != nil {
		t.Fatalf("LastRestart parse failed: %v", err)
	}
	if ts.Before(beforeRestart.Add(-1 * time.Second)) {
		t.Fatalf("LastRestart appears stale: %s", got.LastRestart)
	}

	// Re-open from disk to verify persistence.
	wlReloaded := NewJSONWatchlist(path)
	gotReloaded, err := wlReloaded.Get(ctx, entry.Name)
	if err != nil {
		t.Fatalf("Get(reloaded) returned error: %v", err)
	}
	if gotReloaded.AutoRestart != got.AutoRestart ||
		gotReloaded.RestartCount != got.RestartCount ||
		gotReloaded.FailCount != got.FailCount ||
		gotReloaded.LastRestart != got.LastRestart {
		t.Fatalf("reloaded entry mismatch: got %+v want %+v", gotReloaded, got)
	}
}

func TestTrackedPIDLifecycleAndEphemeralBehavior(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	entry := testEntry("svc-pid")
	if err := wl.Add(ctx, entry); err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	pid, err := wl.GetTrackedPID(ctx, entry.Name)
	if err != nil {
		t.Fatalf("GetTrackedPID() returned error: %v", err)
	}
	if pid != 0 {
		t.Fatalf("GetTrackedPID() = %d, want 0 for unset pid", pid)
	}

	if err := wl.SetTrackedPID(ctx, entry.Name, 12345); err != nil {
		t.Fatalf("SetTrackedPID() returned error: %v", err)
	}
	pid, err = wl.GetTrackedPID(ctx, entry.Name)
	if err != nil {
		t.Fatalf("GetTrackedPID() returned error: %v", err)
	}
	if pid != 12345 {
		t.Fatalf("GetTrackedPID() = %d, want 12345", pid)
	}

	if err := wl.SetTrackedPID(ctx, "missing", 99); err == nil {
		t.Fatalf("SetTrackedPID(missing) expected error")
	}

	if err := wl.Remove(ctx, entry.Name); err != nil {
		t.Fatalf("Remove() returned error: %v", err)
	}

	if _, err := wl.GetTrackedPID(ctx, entry.Name); err == nil {
		t.Fatalf("GetTrackedPID(removed) expected error")
	}

	// Tracked PID data is ephemeral and not loaded from disk.
	entry2 := testEntry("svc-pid-ephemeral")
	if err := wl.Add(ctx, entry2); err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}
	if err := wl.SetTrackedPID(ctx, entry2.Name, 4444); err != nil {
		t.Fatalf("SetTrackedPID() returned error: %v", err)
	}

	wlReloaded := NewJSONWatchlist(path)
	reloadedPID, err := wlReloaded.GetTrackedPID(ctx, entry2.Name)
	if err != nil {
		t.Fatalf("GetTrackedPID(reloaded) returned error: %v", err)
	}
	if reloadedPID != 0 {
		t.Fatalf("GetTrackedPID(reloaded) = %d, want 0", reloadedPID)
	}
}

func TestPersistenceFileContainsExpectedJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	entry := testEntry("svc-json")
	if err := wl.Add(ctx, entry); err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading persisted watchlist: %v", err)
	}

	var parsed []core.WatchlistItem
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("persisted file is not valid json: %v", err)
	}
	if len(parsed) != 1 || parsed[0].Name != entry.Name {
		t.Fatalf("persisted items mismatch: %+v", parsed)
	}
}

func TestMutationReturnsErrorWhenSaveFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	missingParent := filepath.Join(t.TempDir(), "does-not-exist")
	path := filepath.Join(missingParent, "watchlist.json")
	wl := NewJSONWatchlist(path)

	err := wl.Add(ctx, testEntry("svc-save-fail"))
	if err == nil {
		t.Fatalf("Add() expected error when save fails")
	}
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "watchlist.json")
	wl := NewJSONWatchlist(path)

	const workers = 12
	const perWorker = 20

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				name := fmt.Sprintf("svc-%d-%d", w, i)
				entry := testEntry(name)
				if err := wl.Add(ctx, entry); err != nil {
					t.Errorf("Add(%q) error: %v", name, err)
					return
				}
				if _, err := wl.Get(ctx, name); err != nil {
					t.Errorf("Get(%q) error: %v", name, err)
					return
				}
				if err := wl.IncrementFailCount(ctx, name); err != nil {
					t.Errorf("IncrementFailCount(%q) error: %v", name, err)
					return
				}
				if err := wl.ResetFailCount(ctx, name); err != nil {
					t.Errorf("ResetFailCount(%q) error: %v", name, err)
					return
				}
				if err := wl.Update(ctx, name, i%2 == 0); err != nil {
					t.Errorf("Update(%q) error: %v", name, err)
					return
				}
				if err := wl.SetTrackedPID(ctx, name, int32(1000+i)); err != nil {
					t.Errorf("SetTrackedPID(%q) error: %v", name, err)
					return
				}
				if _, err := wl.List(ctx); err != nil {
					t.Errorf("List() error: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()

	items, err := wl.List(ctx)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	want := workers * perWorker
	if len(items) != want {
		t.Fatalf("len(List()) = %d, want %d", len(items), want)
	}
}

func testEntry(name string) core.WatchlistItem {
	return core.WatchlistItem{
		Name:         name,
		RestartCmd:   "echo restart",
		AutoRestart:  true,
		MaxRetries:   3,
		CooldownSecs: 10,
	}
}
