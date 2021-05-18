package exif

import (
	"log"
	"math"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

type exifParser interface {
	ParseExif(media_path string) (*models.MediaEXIF, error)
}

var globalExifParser exifParser

func InitializeEXIFParser() {
	// Decide between internal or external Exif parser
	exiftoolParser, err := newExiftoolParser()

	if err != nil {
		log.Printf("Failed to get exiftool, using internal exif parser instead: %v\n", err)
		globalExifParser = &internalExifParser{}
	} else {
		log.Println("Found exiftool")
		globalExifParser = exiftoolParser
	}
}

// sanitizeFloat overwrites the input value if it's +Inf, -Inf or NaN with a
// valid real number
func sanitizeFloat(v *float64) {
	if math.IsInf(*v, 1) {
		*v = math.MaxFloat64
	} else if math.IsInf(*v, -1) {
		*v = -math.MaxFloat64
	} else if math.IsNaN(*v) {
		*v = 0
	}
}

// sanitizeEXIF overwrites any exif float64 field if it's +Inf, -Inf or Nan
// with a valid real number.
func sanitizeEXIF(exif *models.MediaEXIF) {
	sanitizeFloat(exif.Exposure)
	sanitizeFloat(exif.Aperture)
	sanitizeFloat(exif.FocalLength)
	sanitizeFloat(exif.GPSLatitude)
	sanitizeFloat(exif.GPSLongitude)
}

// SaveEXIF scans the media file for exif metadata and saves it in the database if found
func SaveEXIF(tx *gorm.DB, media *models.Media) (*models.MediaEXIF, error) {

	{
		// Check if EXIF data already exists
		if media.ExifID != nil {

			var exif models.MediaEXIF
			if err := tx.First(&exif, media.ExifID).Error; err != nil {
				return nil, errors.Wrap(err, "get EXIF for media from database")
			}

			return &exif, nil
		}
	}

	if globalExifParser == nil {
		return nil, errors.New("No exif parser initialized")
	}

	exif, err := globalExifParser.ParseExif(media.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse exif data")
	}

	if exif == nil {
		return nil, nil
	}
	sanitizeEXIF(exif)

	// Add EXIF to database and link to media
	if err := tx.Model(&media).Association("Exif").Replace(exif); err != nil {
		return nil, errors.Wrap(err, "save media exif to database")
	}

	if exif.DateShot != nil && !exif.DateShot.Equal(media.DateShot) {
		media.DateShot = *exif.DateShot
		if err := tx.Save(media).Error; err != nil {
			return nil, errors.Wrap(err, "update media date_shot")
		}
	}

	return exif, nil
}
