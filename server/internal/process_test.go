package internal

import (
	"testing"
	"time"
)

func TestParseLogEntry_DownloadProgress(t *testing.T) {
	t.Parallel()

	p := &Process{}
	entry := []byte(`{"percentage":"50.5%","speed":1234567.89,"eta":30}`)

	p.parseLogEntry(entry)

	if p.Progress.Status != StatusDownloading {
		t.Errorf("expected status %d (StatusDownloading), got %d", StatusDownloading, p.Progress.Status)
	}
	if p.Progress.Percentage != "50.5%" {
		t.Errorf("expected percentage %q, got %q", "50.5%", p.Progress.Percentage)
	}
	if p.Progress.Speed != 1234567.89 {
		t.Errorf("expected speed %f, got %f", 1234567.89, p.Progress.Speed)
	}
	if p.Progress.ETA != 30 {
		t.Errorf("expected eta %f, got %f", 30.0, p.Progress.ETA)
	}
}

func TestParseLogEntry_PostprocessFilepath(t *testing.T) {
	t.Parallel()

	p := &Process{}
	entry := []byte(`{"filepath":"/downloads/video.mp4"}`)

	p.parseLogEntry(entry)

	if p.Output.SavedFilePath != "/downloads/video.mp4" {
		t.Errorf("expected filepath %q, got %q", "/downloads/video.mp4", p.Output.SavedFilePath)
	}
}

func TestParseLogEntry_InvalidJSON(t *testing.T) {
	t.Parallel()

	p := &Process{}
	// Should not panic
	p.parseLogEntry([]byte("not json at all"))

	if p.Progress.Status != StatusPending {
		t.Errorf("expected status %d (StatusPending), got %d", StatusPending, p.Progress.Status)
	}
}

func TestParseLogEntry_EmptyJSON(t *testing.T) {
	t.Parallel()

	p := &Process{}
	p.parseLogEntry([]byte(`{}`))

	// Empty JSON still unmarshals successfully into ProgressTemplate,
	// so status gets set to StatusDownloading
	if p.Progress.Status != StatusDownloading {
		t.Errorf("expected status %d (StatusDownloading), got %d", StatusDownloading, p.Progress.Status)
	}
}

func TestParseLogEntry_ProgressThenPostprocess(t *testing.T) {
	t.Parallel()

	p := &Process{}

	p.parseLogEntry([]byte(`{"percentage":"75%","speed":500000,"eta":10}`))
	if p.Progress.Percentage != "75%" {
		t.Errorf("after progress: expected percentage %q, got %q", "75%", p.Progress.Percentage)
	}

	// The postprocess JSON also unmarshals as a valid ProgressTemplate (all zero values),
	// which overwrites the progress fields — this is the actual behavior of parseLogEntry.
	p.parseLogEntry([]byte(`{"filepath":"/downloads/final.mkv"}`))

	if p.Output.SavedFilePath != "/downloads/final.mkv" {
		t.Errorf("expected filepath %q, got %q", "/downloads/final.mkv", p.Output.SavedFilePath)
	}
	// Progress gets reset because postprocess JSON is also valid for ProgressTemplate
	if p.Progress.Status != StatusDownloading {
		t.Errorf("expected status %d (StatusDownloading), got %d", StatusDownloading, p.Progress.Status)
	}
}

func TestSetPending(t *testing.T) {
	t.Parallel()

	p := &Process{Url: "https://example.com/video"}

	before := time.Now()
	p.SetPending()
	after := time.Now()

	if p.Info.URL != "https://example.com/video" {
		t.Errorf("expected URL %q, got %q", "https://example.com/video", p.Info.URL)
	}
	if p.Info.Title != "https://example.com/video" {
		t.Errorf("expected Title to equal URL, got %q", p.Info.Title)
	}
	if p.Progress.Status != StatusPending {
		t.Errorf("expected status %d (StatusPending), got %d", StatusPending, p.Progress.Status)
	}
	if p.Info.CreatedAt.Before(before) || p.Info.CreatedAt.After(after) {
		t.Errorf("expected CreatedAt between %v and %v, got %v", before, after, p.Info.CreatedAt)
	}
}

func TestKill_NilProc(t *testing.T) {
	t.Parallel()

	p := &Process{}
	err := p.Kill()

	if err == nil {
		t.Fatal("expected error for nil proc")
	}
	if err.Error() != "*os.Process not set" {
		t.Errorf("expected error %q, got %q", "*os.Process not set", err.Error())
	}
	// defer in Kill sets status to completed
	if p.Progress.Status != StatusCompleted {
		t.Errorf("expected status %d (StatusCompleted) after kill, got %d", StatusCompleted, p.Progress.Status)
	}
}

func TestBuildFilename_NoExtTemplate(t *testing.T) {
	t.Parallel()

	o := &DownloadOutput{Filename: "myvideo.mp4"}
	buildFilename(o)
	if o.Filename != "myvideo.mp4" {
		t.Errorf("expected %q, got %q", "myvideo.mp4", o.Filename)
	}
}

func TestBuildFilename_WithExtTemplate(t *testing.T) {
	t.Parallel()

	// If filename contains .%(ext)s, buildFilename appends another .%(ext)s
	// then deduplicates
	o := &DownloadOutput{Filename: "myvideo.%(ext)s"}
	buildFilename(o)
	if o.Filename != "myvideo.%(ext)s" {
		t.Errorf("expected %q, got %q", "myvideo.%(ext)s", o.Filename)
	}
}

func TestBuildFilename_EmptyFilename(t *testing.T) {
	t.Parallel()

	o := &DownloadOutput{Filename: ""}
	buildFilename(o)
	if o.Filename != "" {
		t.Errorf("expected empty string, got %q", o.Filename)
	}
}

func TestSanitizeParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  []string
		expect []string
	}{
		{
			name:   "filters shell variable expansion",
			input:  []string{"-f", "best", "${HOME}"},
			expect: []string{"-f", "best"},
		},
		{
			name:   "filters command chaining",
			input:  []string{"-f", "best", "&& rm -rf /"},
			expect: []string{"-f", "best"},
		},
		{
			name:   "filters empty strings",
			input:  []string{"-f", "", "best", ""},
			expect: []string{"-f", "best"},
		},
		{
			name:   "clean params unchanged",
			input:  []string{"-f", "best", "--no-playlist"},
			expect: []string{"-f", "best", "--no-playlist"},
		},
		{
			name:   "nil input",
			input:  nil,
			expect: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := sanitizeParams(tc.input)
			if len(got) != len(tc.expect) {
				t.Fatalf("expected %d params, got %d: %v", len(tc.expect), len(got), got)
			}
			for i := range tc.expect {
				if got[i] != tc.expect[i] {
					t.Errorf("param[%d]: expected %q, got %q", i, tc.expect[i], got[i])
				}
			}
		})
	}
}
