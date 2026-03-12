package internal

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"
	"time"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/common"
)

func TestProgressTemplate_Unmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
		want ProgressTemplate
	}{
		{
			name: "all fields",
			json: `{"percentage":"45.2%","speed":1048576.5,"size":"100MiB","eta":120}`,
			want: ProgressTemplate{Percentage: "45.2%", Speed: 1048576.5, Size: "100MiB", Eta: 120},
		},
		{
			name: "null values",
			json: `{"percentage":"0%","speed":null,"eta":null}`,
			want: ProgressTemplate{Percentage: "0%", Speed: 0, Eta: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got ProgressTemplate
			if err := json.Unmarshal([]byte(tc.json), &got); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if got.Percentage != tc.want.Percentage {
				t.Errorf("Percentage: expected %q, got %q", tc.want.Percentage, got.Percentage)
			}
			if got.Speed != tc.want.Speed {
				t.Errorf("Speed: expected %f, got %f", tc.want.Speed, got.Speed)
			}
			if got.Size != tc.want.Size {
				t.Errorf("Size: expected %q, got %q", tc.want.Size, got.Size)
			}
			if got.Eta != tc.want.Eta {
				t.Errorf("Eta: expected %f, got %f", tc.want.Eta, got.Eta)
			}
		})
	}
}

func TestPostprocessTemplate_Unmarshal(t *testing.T) {
	t.Parallel()

	data := `{"filepath":"/downloads/video.mp4"}`
	var pp PostprocessTemplate
	if err := json.Unmarshal([]byte(data), &pp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if pp.FilePath != "/downloads/video.mp4" {
		t.Errorf("expected filepath %q, got %q", "/downloads/video.mp4", pp.FilePath)
	}
}

func TestDownloadRequest_Unmarshal(t *testing.T) {
	t.Parallel()

	data := `{"url":"https://example.com","path":"/tmp","rename":"out.mp4","params":["-f","best"]}`
	var dr DownloadRequest
	if err := json.Unmarshal([]byte(data), &dr); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if dr.URL != "https://example.com" {
		t.Errorf("expected URL %q, got %q", "https://example.com", dr.URL)
	}
	if dr.Path != "/tmp" {
		t.Errorf("expected path %q, got %q", "/tmp", dr.Path)
	}
	if dr.Rename != "out.mp4" {
		t.Errorf("expected rename %q, got %q", "out.mp4", dr.Rename)
	}
	if len(dr.Params) != 2 || dr.Params[0] != "-f" || dr.Params[1] != "best" {
		t.Errorf("expected params [-f best], got %v", dr.Params)
	}
}

func TestSession_GobRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Second)

	original := Session{
		Processes: []ProcessResponse{
			{
				Id: "test-id-1",
				Progress: DownloadProgress{
					Status:     StatusCompleted,
					Percentage: "-1",
					Speed:      0,
					ETA:        0,
				},
				Info: common.DownloadInfo{
					URL:       "https://example.com/video",
					Title:     "Test Video",
					Thumbnail: "https://example.com/thumb.jpg",
					CreatedAt: now,
				},
				Output: DownloadOutput{
					Path:          "/downloads",
					Filename:      "test.mp4",
					SavedFilePath: "/downloads/test.mp4",
				},
				Params: []string{"-f", "best"},
			},
		},
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(original); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	var decoded Session
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(decoded.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(decoded.Processes))
	}

	got := decoded.Processes[0]
	if got.Id != "test-id-1" {
		t.Errorf("expected id %q, got %q", "test-id-1", got.Id)
	}
	if got.Info.Title != "Test Video" {
		t.Errorf("expected title %q, got %q", "Test Video", got.Info.Title)
	}
	if !got.Info.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, got.Info.CreatedAt)
	}
	if got.Output.SavedFilePath != "/downloads/test.mp4" {
		t.Errorf("expected saved path %q, got %q", "/downloads/test.mp4", got.Output.SavedFilePath)
	}
	if len(got.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(got.Params))
	}
}

func TestSession_GobRoundTrip_Empty(t *testing.T) {
	t.Parallel()

	original := Session{Processes: []ProcessResponse{}}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(original); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	var decoded Session
	if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(decoded.Processes) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(decoded.Processes))
	}
}
