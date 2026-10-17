package processing_tasks

import (
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/pkg/errors"
)

type ProcessPhotoTask struct {
	scanner_task.ScannerTaskBase
}

func (t ProcessPhotoTask) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) ([]*models.MediaURL, error) {
	if mediaData.Media.Type != models.MediaTypePhoto {
		return []*models.MediaURL{}, nil
	}

	updatedURLs := make([]*models.MediaURL, 0)
	photo := mediaData.Media

	log.Info(ctx, "Processing photo", "photo", photo.Path)

	photoURLFromDB := makePhotoURLChecker(ctx.GetDB(), photo.ID)

	// original photo url
	origURL, err := photoURLFromDB(models.MediaOriginal)
	if err != nil {
		return []*models.MediaURL{}, err
	}

	// Thumbnail
	thumbURL, err := photoURLFromDB(models.PhotoThumbnail)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing photo thumbnail")
	}

	// Highres
	highResURL, err := photoURLFromDB(models.PhotoHighRes)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing photo highres")
	}

	var baseImagePath string = photo.Path

	// Generate high res jpeg
	if highResURL == nil {

		contentType, err := mediaData.ContentType()
		if err != nil {
			return []*models.MediaURL{}, err
		}

		if !contentType.IsWebCompatible() {
			highresName := generateUniqueMediaNamePrefixed("highres", photo.Path, ".jpg")
			baseImagePath = path.Join(mediaCachePath, highresName)

			highRes, err := generateSaveHighResJPEG(ctx.GetDB(), photo, mediaData, highresName, baseImagePath, nil)
			if err != nil {
				return []*models.MediaURL{}, err
			}

			updatedURLs = append(updatedURLs, highRes)
		}
	} else {
		// Verify that highres photo still exists in cache
		baseImagePath = path.Join(mediaCachePath, highResURL.MediaName)

		if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
			log.Info(ctx, "High-res photo found in database but not in cache, re-encoding photo to cache", "media_name", highResURL.MediaName)
			updatedURLs = append(updatedURLs, highResURL)

			err = mediaData.EncodeHighRes(baseImagePath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "creating high-res cached image")
			}
		}
	}

	// Save original photo to database
	if origURL == nil {

		// Make sure photo dimensions is set
		photoDimensions, err := media_encoding.GetPhotoDimensions(photo.Path)
		if err != nil {
			return []*models.MediaURL{}, err
		}

		original, err := saveOriginalPhotoToDB(ctx.GetDB(), photo, mediaData, photoDimensions)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "saving original photo to database")
		}

		updatedURLs = append(updatedURLs, original)
	}

	// Save thumbnail to cache
	if thumbURL == nil {
		thumbnailName := generateUniqueMediaNamePrefixed("thumbnail", photo.Path, ".jpg")
		thumbnail, err := generateSaveThumbnailJPEG(ctx.GetDB(), photo, thumbnailName, mediaCachePath, baseImagePath, nil)
		if err != nil {
			return []*models.MediaURL{}, err
		}

		updatedURLs = append(updatedURLs, thumbnail)
	} else {
		// Verify that thumbnail photo still exists in cache
		thumbPath := path.Join(mediaCachePath, thumbURL.MediaName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			updatedURLs = append(updatedURLs, thumbURL)
			log.Info(ctx, "Thumbnail photo found in database but not in cache, re-encoding photo to cache", "media_name", thumbURL.MediaName)

			_, err := media_encoding.EncodeThumbnail(ctx.GetDB(), baseImagePath, thumbPath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "could not create thumbnail cached image")
			}
		}
	}

	// Process Live Photo companion video if present
	if liveVideoPath, ok := media_type.FindLivePhotoVideoCounterpart(photo.Path); ok {
		if photo.LivePhotoPath == nil || *photo.LivePhotoPath != liveVideoPath {
			photo.LivePhotoPath = &liveVideoPath
			if err := ctx.GetDB().Model(photo).Update("live_photo_path", liveVideoPath).Error; err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "updating photo live_photo_path in database")
			}
		}

		livePhotoURL, err := photoURLFromDB(models.LivePhotoVideo)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "error checking photo live_photo_video url")
		}

		if livePhotoURL == nil {
			fileStats, err := os.Stat(liveVideoPath)
			if err == nil {
				mediaName := generateUniqueMediaNamePrefixed("live_video", liveVideoPath, path.Ext(liveVideoPath))
				videoType := media_type.GetMediaType(liveVideoPath)
				contentType := videoType.String()
				if contentType == "unknown" {
					contentType = "video/mp4"
				}

				liveMediaURL := models.MediaURL{
					MediaID:     photo.ID,
					MediaName:   mediaName,
					Width:       0,
					Height:      0,
					Purpose:     models.LivePhotoVideo,
					ContentType: contentType,
					FileSize:    fileStats.Size(),
				}

				if err := ctx.GetDB().Create(&liveMediaURL).Error; err != nil {
					return []*models.MediaURL{}, errors.Wrap(err, "saving live photo url to database")
				}

				updatedURLs = append(updatedURLs, &liveMediaURL)
			}
		}
	} else if photo.LivePhotoPath != nil {
		photo.LivePhotoPath = nil
		if err := ctx.GetDB().Model(photo).Update("live_photo_path", nil).Error; err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "clearing photo live_photo_path in database")
		}
		if err := ctx.GetDB().Where("media_id = ? AND purpose = ?", photo.ID, models.LivePhotoVideo).Delete(&models.MediaURL{}).Error; err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "deleting old live photo video urls from database")
		}
	}

	return updatedURLs, nil
}
