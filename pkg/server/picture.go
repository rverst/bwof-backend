package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/rs/zerolog/log"
	"github.com/rverst/bwof-backend/pkg/helper"
	"github.com/rverst/bwof-backend/pkg/models"
	"goji.io/pat"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

type editBody struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type disableBody struct {
	Disable bool `json:"disable"`
}

type cropBody struct {
	Crop crop `json:"crop"`
}

type crop struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func getPicture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}

	pic, err := getDbPicture(id)
	if err != nil {
		res := models.Response{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}
		_, _ = helper.WriteJson(w, res.Status, res)
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*pic))
}

func getPictures(w http.ResponseWriter, _ *http.Request) {
	pics, err := getDbPictures()
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusNotFound, "")
		return
	}

	if len(pics) == 0 {
		_, _ = helper.WriteError(w, http.StatusNotFound, "no pictures found in database")
		return
	}

	sort.Slice(pics, func(i, j int) bool {
		return pics[i].Uploaded.After(pics[j].Uploaded)
	})

	list := make([]pictureResponse, len(pics))
	for i, p := range pics {
		list[i] = fromPicture(p)
	}

	_, _ = helper.WriteJson(w, http.StatusOK, list)
}

func cropPicture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}

	var body cropBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info().Str("id", id.String()).Interface("body", body).Msg("cropPicture")

	pic, err := getDbPicture(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	oW := pic.OriginalBounds.Dx()
	oH := pic.OriginalBounds.Dy()

	body.Crop.Width = min(body.Crop.Width, oW)
	body.Crop.Height = min(body.Crop.Height, oH)

	log.Info().Interface("crop", body.Crop).Msg("crop")

	if body.Crop.Width == 0 || body.Crop.Height == 0 ||
		(body.Crop.Width >= oW && body.Crop.Height >= oH) {
		pic.UseCropped = false
		pic.Edited = time.Now()

		err = updatePicture(pic)
		if err != nil {
			_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info().Msg("crop disabled")
		_, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*pic))
		return
	}

	dir := path.Join(pictureDir, id.String())
	ext := filepath.Ext(pic.OriginalPath)
	fileName := fmt.Sprintf("crop%s", ext)
	thumbName := fmt.Sprintf("thumb_crop%s", ext)
	cropName := path.Join(dir, fileName)
	thumbCropName := path.Join(dir, thumbName)

	f, _ := os.Open(path.Join(dir, pic.OriginalPath))
	img, _, err := image.Decode(f)
	if err != nil {
		log.Error().Err(err).Send()
	}
	defer helper.Close(f, pic.OriginalPath)

	x0 := body.Crop.X
	y0 := body.Crop.Y
	x1 := body.Crop.X + body.Crop.Width
	y1 := body.Crop.Y + body.Crop.Height

	if x0 < 0 {
		t := x0 * -1
		x0 = 0
		if (x1 + t) <= img.Bounds().Max.X {
			x1 += t
		}
	}
	if y0 < 0 {
		t := y0 * -1
		y0 = 0
		if (y1 + t) <= img.Bounds().Max.Y {
			y1 += t
		}
	}

	cb := image.Rect(x0, y0, x1, y1)
	log.Info().Interface("bounds", cb).Msg("crop")
	cropped, err := cutter.Crop(img, cutter.Config{
		Width:  x1 - x0,
		Height: y1 - y0,
		Anchor: image.Point{
			X: x0,
			Y: y0,
		},
		Options: cutter.Copy,
	})

	cf, _ := os.OpenFile(cropName, os.O_WRONLY|os.O_CREATE, 0660)
	defer helper.Close(cf, cropName)

	ct, _ := os.OpenFile(thumbCropName, os.O_WRONLY|os.O_CREATE, 0660)
	defer helper.Close(cf, thumbCropName)

	thumb := resize.Thumbnail(helper.ThumbnailSize, helper.ThumbnailSize, cropped, resize.Lanczos3)

	if ext == ".png" {
		_ = png.Encode(cf, cropped)
		_ = png.Encode(ct, thumb)
	} else if ext == ".jpg" {
		o := &jpeg.Options{Quality: 100}
		_ = jpeg.Encode(cf, cropped, o)
		_ = jpeg.Encode(ct, thumb, o)
	}

	pic.CroppedBounds = cb
	pic.CroppedUrl = fmt.Sprintf("/pictures/%s/%s", pic.Id.String(), filepath.Base(cropName))
	pic.CroppedPath = cropName
	pic.ThumbCroppedUrl = fmt.Sprintf("/pictures/%s/%s", pic.Id.String(), filepath.Base(thumbCropName))
	pic.ThumbCroppedPath = thumbCropName
	pic.UseCropped = true
	pic.Edited = time.Now()

	err = updatePicture(pic)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*pic))
}

func editPictureContent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}

	var body editBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info().Str("id", id.String()).Interface("body", body).Msg("editPictureContent")

	if !plainTextRegex.MatchString(body.Title) ||
		!plainTextRegex.MatchString(body.Text) {
		err := fmt.Errorf("only plain text allowed in title/text")
		log.Warn().Err(err).Msg("editPictureContent")
		_, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	pic, err := getDbPicture(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pic.Content.Title = body.Title
	pic.Content.Text = body.Text
	err = updatePicture(pic)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*pic))
}

func disablePicture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}
	var body disableBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Info().Str("id", id.String()).Interface("body", body).Msg("disablePicture")

	pic, err := getDbPicture(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pic.Disabled = body.Disable
	err = updatePicture(pic)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*pic))
}

func deletePicture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}
	log.Info().Str("id", id.String()).Msg("deletePicture")

	pic, err := getDbPicture(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = deleteDbPicture(pic.Id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p := path.Join(pictureDir, pic.Id.String())
	err = os.RemoveAll(p)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func min(x1 int, x2 int) int {
	if x1 < x2 {
		return x1
	}
	return x2
}
