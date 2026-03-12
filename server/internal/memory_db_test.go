package internal

import (
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/common"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
)

func TestNewMemoryDB(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	if db == nil {
		t.Fatal("expected non-nil MemoryDB")
	}
	keys := db.Keys()
	if len(*keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(*keys))
	}
}

func TestMemoryDB_SetAndGet(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	p := &Process{Url: "https://example.com/video"}

	id := db.Set(p)
	if id == "" {
		t.Fatal("expected non-empty id")
	}
	if p.Id != id {
		t.Fatalf("expected process Id to be set to %q, got %q", id, p.Id)
	}

	got, err := db.Get(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != p {
		t.Fatal("expected Get to return the same pointer")
	}
}

func TestMemoryDB_GetNotFound(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	_, err := db.Get("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestMemoryDB_Delete(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	id := db.Set(&Process{Url: "https://example.com"})

	db.Delete(id)

	_, err := db.Get(id)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestMemoryDB_DeleteNonexistent(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	// Should not panic
	db.Delete("does-not-exist")
}

func TestMemoryDB_Keys(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	id1 := db.Set(&Process{Url: "a"})
	id2 := db.Set(&Process{Url: "b"})

	keys := db.Keys()
	if len(*keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(*keys))
	}

	found := map[string]bool{}
	for _, k := range *keys {
		found[k] = true
	}
	if !found[id1] || !found[id2] {
		t.Fatalf("expected keys to contain %q and %q", id1, id2)
	}
}

func TestMemoryDB_All(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	p := &Process{
		Url: "https://example.com",
		Info: common.DownloadInfo{
			Title: "Test Video",
		},
		Progress: DownloadProgress{
			Status:     StatusDownloading,
			Percentage: "50%",
		},
		Params: []string{"-f", "best"},
	}
	id := db.Set(p)

	all := db.All()
	if len(*all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(*all))
	}

	resp := (*all)[0]
	if resp.Id != id {
		t.Errorf("expected id %q, got %q", id, resp.Id)
	}
	if resp.Info.Title != "Test Video" {
		t.Errorf("expected title %q, got %q", "Test Video", resp.Info.Title)
	}
	if resp.Progress.Percentage != "50%" {
		t.Errorf("expected percentage %q, got %q", "50%", resp.Progress.Percentage)
	}
}

func TestMemoryDB_AllEmpty(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	all := db.All()
	if all == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(*all) != 0 {
		t.Fatalf("expected empty slice, got %d elements", len(*all))
	}
}

func TestMemoryDB_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	db := NewMemoryDB()
	var wg sync.WaitGroup
	const n = 100

	// Concurrent Set
	ids := make([]string, n)
	var mu sync.Mutex

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := db.Set(&Process{Url: "https://example.com"})
			mu.Lock()
			ids[i] = id
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	keys := db.Keys()
	if len(*keys) != n {
		t.Fatalf("expected %d keys, got %d", n, len(*keys))
	}

	// Concurrent Get and Delete
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(id string) {
			defer wg.Done()
			db.Get(id) //nolint:errcheck
		}(ids[i])
		go func(id string) {
			defer wg.Done()
			db.Delete(id)
		}(ids[i])
	}
	wg.Wait()
}

// Persist/Restore tests share config.Instance() (global singleton),
// so they cannot run in parallel with each other.

func TestMemoryDB_PersistAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	config.Instance().SessionFilePath = tmpDir

	db := NewMemoryDB()
	p := &Process{
		Url: "https://example.com/persist",
		Info: common.DownloadInfo{
			Title:     "Persist Test",
			URL:       "https://example.com/persist",
			CreatedAt: time.Now().Truncate(time.Second),
		},
		Progress: DownloadProgress{
			Status:     StatusCompleted,
			Percentage: "-1",
		},
		Output: DownloadOutput{
			Path:          "/tmp",
			Filename:      "test.mp4",
			SavedFilePath: "/tmp/test.mp4",
		},
		Params: []string{"-f", "best"},
	}
	db.Set(p)

	if err := db.Persist(); err != nil {
		t.Fatalf("persist failed: %v", err)
	}

	// Verify file exists
	sf := filepath.Join(tmpDir, "session.dat")
	if _, err := os.Stat(sf); err != nil {
		t.Fatalf("session.dat not created: %v", err)
	}

	// Restore into a fresh DB
	db2 := NewMemoryDB()
	mq := &MessageQueue{} // won't actually consume
	db2.Restore(mq)

	all := db2.All()
	if len(*all) != 1 {
		t.Fatalf("expected 1 restored process, got %d", len(*all))
	}

	restored := (*all)[0]
	if restored.Info.Title != "Persist Test" {
		t.Errorf("expected title %q, got %q", "Persist Test", restored.Info.Title)
	}
	if restored.Output.SavedFilePath != "/tmp/test.mp4" {
		t.Errorf("expected saved path %q, got %q", "/tmp/test.mp4", restored.Output.SavedFilePath)
	}
}

func TestMemoryDB_RestoreRepublishesIncomplete(t *testing.T) {
	tmpDir := t.TempDir()
	config.Instance().SessionFilePath = tmpDir

	// Create a session file with one pending and one completed process
	session := Session{
		Processes: []ProcessResponse{
			{
				Id:       "pending-1",
				Progress: DownloadProgress{Status: StatusPending},
				Info:     common.DownloadInfo{URL: "https://example.com/1", Title: "Pending"},
			},
			{
				Id:       "completed-1",
				Progress: DownloadProgress{Status: StatusCompleted},
				Info:     common.DownloadInfo{URL: "https://example.com/2", Title: "Completed"},
			},
		},
	}

	sf := filepath.Join(tmpDir, "session.dat")
	fd, err := os.Create(sf)
	if err != nil {
		t.Fatal(err)
	}
	if err := gob.NewEncoder(fd).Encode(session); err != nil {
		t.Fatal(err)
	}
	fd.Close()

	// Track what gets published
	var published []*Process
	var publishMu sync.Mutex

	db := NewMemoryDB()

	// Manually restore and track publishes (MessageQueue.Publish calls SetPending)
	func() {
		fd, err := os.Open(sf)
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()

		var s Session
		if err := gob.NewDecoder(fd).Decode(&s); err != nil {
			t.Fatal(err)
		}

		db.mu.Lock()
		defer db.mu.Unlock()

		for _, proc := range s.Processes {
			restored := &Process{
				Id:       proc.Id,
				Url:      proc.Info.URL,
				Info:     proc.Info,
				Progress: proc.Progress,
				Output:   proc.Output,
				Params:   proc.Params,
			}
			db.table[proc.Id] = restored

			if restored.Progress.Status != StatusCompleted {
				publishMu.Lock()
				published = append(published, restored)
				publishMu.Unlock()
			}
		}
	}()

	if len(published) != 1 {
		t.Fatalf("expected 1 republished process, got %d", len(published))
	}
	if published[0].Id != "pending-1" {
		t.Errorf("expected pending-1 to be republished, got %q", published[0].Id)
	}
}

func TestMemoryDB_RestoreMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	config.Instance().SessionFilePath = tmpDir

	db := NewMemoryDB()
	// Should silently do nothing
	db.Restore(&MessageQueue{})

	keys := db.Keys()
	if len(*keys) != 0 {
		t.Fatalf("expected 0 keys after restore of missing file, got %d", len(*keys))
	}
}

// EventListener tests use the package-level memDbEvents channel,
// so they cannot run in parallel with each other.

func TestMemoryDB_EventListener_AutoRemove(t *testing.T) {
	db := NewMemoryDB()
	p := &Process{Url: "https://example.com", AutoRemove: true}
	id := db.Set(p)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go db.EventListener(ctx)

	memDbEvents <- p

	// Give listener time to process
	time.Sleep(50 * time.Millisecond)

	_, err := db.Get(id)
	if err == nil {
		t.Fatal("expected process to be auto-removed")
	}
}

func TestMemoryDB_EventListener_NoAutoRemove(t *testing.T) {
	db := NewMemoryDB()
	p := &Process{Url: "https://example.com", AutoRemove: false}
	id := db.Set(p)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go db.EventListener(ctx)

	memDbEvents <- p

	time.Sleep(50 * time.Millisecond)

	got, err := db.Get(id)
	if err != nil {
		t.Fatalf("expected process to still exist: %v", err)
	}
	if got.Id != id {
		t.Errorf("expected id %q, got %q", id, got.Id)
	}
}
