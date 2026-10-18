package media_type

import (
	"os"
	"path"
	"strings"
)

var (
	livePhotoVideoExtensions = []string{
		".mov", ".MOV", ".mp4", ".MP4", ".m4v", ".M4V",
	}

	livePhotoImageExtensions = []string{
		".heic", ".HEIC", ".heif", ".HEIF",
		".jpg", ".JPG", ".jpeg", ".JPEG",
		".png", ".PNG", ".dng", ".DNG",
		".webp", ".WEBP",
	}

	webCompatibleExtensions = []string{
		".jpg", ".JPG", ".jpeg", ".JPEG",
		".png", ".PNG", ".webp", ".WEBP",
	}

	rawExtensions = []string{
		".tiff", ".TIFF", ".tif", ".TIF",
		".dng", ".DNG", ".cr2", ".CR2", ".cr3", ".CR3",
		".nef", ".NEF", ".arw", ".ARW", ".orf", ".ORF",
		".rw2", ".RW2", ".pef", ".PEF", ".raf", ".RAF",
	}
)

func findDirectCounterpart(filename string, candidateExts []string) (string, bool) {
	ext := path.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	for _, candExt := range candidateExts {
		if strings.EqualFold(ext, candExt) {
			continue
		}
		candPath := base + candExt
		if fi, err := os.Stat(candPath); err == nil && !fi.IsDir() {
			return candPath, true
		}
	}
	return "", false
}

// FindWebCounterpart returns the filename if the file imagePath has a counterpart file compatible with the browser.
func FindWebCounterpart(imagePath string) (string, bool) {
	return findDirectCounterpart(imagePath, webCompatibleExtensions)
}

// FindRawCounterpart returns the filename if the file imagePath has a counterpart file which needs to be processed before showing in the browser.
func FindRawCounterpart(imagePath string) (string, bool) {
	return findDirectCounterpart(imagePath, rawExtensions)
}

// FindLivePhotoVideoCounterpart returns the companion video path if the file imagePath has a companion video (e.g. .mov, .mp4).
func FindLivePhotoVideoCounterpart(imagePath string) (string, bool) {
	return findDirectCounterpart(imagePath, livePhotoVideoExtensions)
}

// FindLivePhotoImageCounterpart returns the companion image path if the file videoPath is a companion video for an image (e.g. .heic, .jpg).
func FindLivePhotoImageCounterpart(videoPath string) (string, bool) {
	return findDirectCounterpart(videoPath, livePhotoImageExtensions)
}
