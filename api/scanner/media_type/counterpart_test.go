package media_type

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

func TestFindWebCounterpart(t *testing.T) {
	mediaPath := test_utils.PathFromAPIRoot("scanner", "test_media", "real_media")

	tests := []struct {
		input    string
		wantFile string
		wantOk   bool
	}{
		{"raw_with_jpg.tiff", "raw_with_jpg.jpg", true},
		{"raw_with_file.tiff", "", false},
		{"standalone_raw.tiff", "", false},
	}

	for _, tc := range tests {
		input := filepath.Join(mediaPath, tc.input)
		if _, err := os.Stat(input); err != nil {
			t.Fatalf("input %q doesn't exist: %v", input, err)
		}

		file, ok := FindWebCounterpart(input)
		got := strings.TrimLeft(strings.TrimPrefix(file, mediaPath), "/")

		if got != tc.wantFile || ok != tc.wantOk {
			t.Errorf("FindWebCounterpart(%q) = (%q, %v), want: (%q, %v)", tc.input, got, ok, tc.wantFile, tc.wantOk)
		}
	}
}

func TestFindRawCounterpart(t *testing.T) {
	mediaPath := test_utils.PathFromAPIRoot("scanner", "test_media", "real_media")

	tests := []struct {
		input    string
		wantFile string
		wantOk   bool
	}{
		{"raw_with_jpg.jpg", "raw_with_jpg.tiff", true},
		{"jpg_with_file.jpg", "", false},
		{"standalone_jpg.jpg", "", false},
	}

	for _, tc := range tests {
		input := filepath.Join(mediaPath, tc.input)
		if _, err := os.Stat(input); err != nil {
			t.Fatalf("input %q doesn't exist: %v", input, err)
		}

		file, ok := FindRawCounterpart(input)
		got := strings.TrimLeft(strings.TrimPrefix(file, mediaPath), "/")

		if got != tc.wantFile || ok != tc.wantOk {
			t.Errorf("FindRawCounterpart(%q) = (%q, %v), want: (%q, %v)", tc.input, got, ok, tc.wantFile, tc.wantOk)
		}
	}
}

func TestFindLivePhotoCounterpart(t *testing.T) {
	tmpDir := t.TempDir()
	photoPath := filepath.Join(tmpDir, "IMG_1001.HEIC")
	videoPath := filepath.Join(tmpDir, "IMG_1001.MOV")

	if err := os.WriteFile(photoPath, []byte("fake-heic"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoPath, []byte("fake-mov"), 0644); err != nil {
		t.Fatal(err)
	}

	foundVideo, ok := FindLivePhotoVideoCounterpart(photoPath)
	if !ok || foundVideo != videoPath {
		t.Errorf("FindLivePhotoVideoCounterpart(%q) = (%q, %v), want (%q, true)", photoPath, foundVideo, ok, videoPath)
	}

	foundPhoto, ok := FindLivePhotoImageCounterpart(videoPath)
	if !ok || foundPhoto != photoPath {
		t.Errorf("FindLivePhotoImageCounterpart(%q) = (%q, %v), want (%q, true)", videoPath, foundPhoto, ok, photoPath)
	}
}

