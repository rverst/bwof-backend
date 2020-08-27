package models

import (
	"github.com/google/uuid"
	"image"
	"time"
)

type Picture struct {
	Id               uuid.UUID       `json:"id"`
	Content          Content         `json:"content"`
	CroppedBounds    image.Rectangle `json:"cropped_bounds"`
	CroppedPath      string          `json:"crop_path"`
	CroppedUrl       string          `json:"cropped_url"`
	Disabled         bool            `json:"disabled"`
	Edited           time.Time       `json:"edited"`
	OriginalBounds   image.Rectangle `json:"original_bounds"`
	OriginalPath     string          `json:"original_path"`
	OriginalUrl      string          `json:"original_url"`
	ThumbnailPath    string          `json:"thumbnail_path"`
	ThumbnailUrl     string          `json:"thumbnail_url"`
	ThumbCroppedPath string          `json:"thumb_crop_path"`
	ThumbCroppedUrl  string          `json:"thumb_cropped_url"`
	TopCrop          image.Rectangle `json:"top_crop"`
	Uploaded         time.Time       `json:"uploaded"`
	UploadedFilename string          `json:"uploaded_filename"`
	Uploader         string          `json:"uploader"`
	UseCropped       bool            `json:"useCropped"`
}

type Content struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
