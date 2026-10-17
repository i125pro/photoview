package media_type

import (
	"path"
	"path/filepath"
	"strings"
)

// FindWebCounterpart returns the filename if the file `imagePath` has a counterpart file competible with the browser.
func FindWebCounterpart(imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
		return mt.IsImage() && mt.IsWebCompatible()
	})
}

// FindRawCounterpart returns the filename if the file `imagePath` has a counterpart file which needs to be processed before showing in the browser.
func FindRawCounterpart(imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
		return mt.IsImage() && !mt.IsWebCompatible()
	})
}

// FindLivePhotoVideoCounterpart returns the companion video path if the file `imagePath` has a companion video (e.g. .mov, .mp4).
func FindLivePhotoVideoCounterpart(imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		ext := strings.ToLower(path.Ext(filename))
		if ext == ".mov" || ext == ".mp4" || ext == ".m4v" {
			return true
		}
		mt := GetMediaType(filename)
		return mt.IsVideo()
	})
}

// FindLivePhotoImageCounterpart returns the companion image path if the file `videoPath` is a companion video for an image (e.g. .heic, .jpg).
func FindLivePhotoImageCounterpart(videoPath string) (string, bool) {
	return findCounterpart(videoPath, func(filename string) bool {
		ext := strings.ToLower(path.Ext(filename))
		if ext == ".heic" || ext == ".heif" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".dng" {
			return true
		}
		mt := GetMediaType(filename)
		return mt.IsImage()
	})
}

func findCounterpart(filename string, acceptFn func(filepath string) bool) (string, bool) {
	ext := path.Ext(filename)
	filenamePattern := strings.TrimSuffix(filename, ext) + ".*"

	filenames, err := filepath.Glob(filenamePattern)
	if err != nil {
		return "", false
	}

	for _, f := range filenames {
		if strings.EqualFold(f, filename) {
			continue
		}

		if acceptFn(f) {
			return f, true
		}
	}

	return "", false
}
