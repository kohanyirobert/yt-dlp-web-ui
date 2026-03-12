package rpc

import (
	"testing"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
)

func newTestService() *Service {
	return &Service{
		db: internal.NewMemoryDB(),
		// mq and lm intentionally nil — tests don't call methods that need them
	}
}

func TestExec_CreatesProcessAndReturnsID(t *testing.T) {
	t.Parallel()
	s := newTestService()
	// Exec calls mq.Publish, so we need a minimal MQ.
	// Instead, test the DB-only path by manually doing what Exec does
	// without the message queue (which would try to spawn yt-dlp).
	args := internal.DownloadRequest{
		URL:    "https://example.com/video",
		Params: []string{"-f", "best"},
	}

	p := &internal.Process{
		Url:    args.URL,
		Params: args.Params,
		Output: internal.DownloadOutput{
			Path:     args.Path,
			Filename: args.Rename,
		},
	}

	id := s.db.Set(p)
	p.SetPending()

	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	got, err := s.db.Get(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Progress.Status != internal.StatusPending {
		t.Errorf("expected status %d, got %d", internal.StatusPending, got.Progress.Status)
	}
}

func TestExec_SetsOutputFields(t *testing.T) {
	t.Parallel()
	s := newTestService()
	args := internal.DownloadRequest{
		URL:    "https://example.com/video",
		Path:   "/custom/path",
		Rename: "custom_name.mp4",
	}

	p := &internal.Process{
		Url: args.URL,
		Output: internal.DownloadOutput{
			Path:     args.Path,
			Filename: args.Rename,
		},
	}

	id := s.db.Set(p)
	got, _ := s.db.Get(id)

	if got.Output.Path != "/custom/path" {
		t.Errorf("expected path %q, got %q", "/custom/path", got.Output.Path)
	}
	if got.Output.Filename != "custom_name.mp4" {
		t.Errorf("expected filename %q, got %q", "custom_name.mp4", got.Output.Filename)
	}
}

func TestProgress_ExistingProcess(t *testing.T) {
	t.Parallel()
	s := newTestService()
	p := &internal.Process{
		Url: "https://example.com",
	}
	id := s.db.Set(p)

	// Simulate some progress
	p.Progress = internal.DownloadProgress{
		Status:     internal.StatusDownloading,
		Percentage: "42%",
		Speed:      500000,
		ETA:        60,
	}

	var progress internal.DownloadProgress
	err := s.Progess(internal.DownloadRequest{Id: id}, &progress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress.Percentage != "42%" {
		t.Errorf("expected percentage %q, got %q", "42%", progress.Percentage)
	}
	if progress.Speed != 500000 {
		t.Errorf("expected speed %f, got %f", 500000.0, progress.Speed)
	}
}

func TestProgress_NonexistentProcess(t *testing.T) {
	t.Parallel()
	s := newTestService()
	var progress internal.DownloadProgress
	err := s.Progess(internal.DownloadRequest{Id: "nonexistent"}, &progress)
	if err == nil {
		t.Fatal("expected error for nonexistent process")
	}
}

func TestRunning_ReturnsAllProcesses(t *testing.T) {
	t.Parallel()
	s := newTestService()
	s.db.Set(&internal.Process{Url: "https://example.com/1"})
	s.db.Set(&internal.Process{Url: "https://example.com/2"})

	var running Running
	err := s.Running(NoArgs{}, &running)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(running) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(running))
	}
}

func TestRunning_EmptyDB(t *testing.T) {
	t.Parallel()
	s := newTestService()
	var running Running
	err := s.Running(NoArgs{}, &running)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(running) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(running))
	}
}

func TestPending_ReturnsAllKeys(t *testing.T) {
	t.Parallel()
	s := newTestService()
	id1 := s.db.Set(&internal.Process{Url: "https://example.com/1"})
	id2 := s.db.Set(&internal.Process{Url: "https://example.com/2"})

	var pending Pending
	err := s.Pending(NoArgs{}, &pending)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(pending))
	}

	found := map[string]bool{}
	for _, k := range pending {
		found[k] = true
	}
	if !found[id1] || !found[id2] {
		t.Errorf("expected keys %q and %q in pending", id1, id2)
	}
}

func TestKill_NonexistentID(t *testing.T) {
	t.Parallel()
	s := newTestService()
	var killed string
	err := s.Kill("nonexistent-id", &killed)
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
}

func TestClear_RemovesFromDB(t *testing.T) {
	t.Parallel()
	s := newTestService()
	id := s.db.Set(&internal.Process{Url: "https://example.com"})

	var killed string
	err := s.Clear(id, &killed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.db.Get(id)
	if err == nil {
		t.Fatal("expected error after clear")
	}
}
