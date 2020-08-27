package server

import (
  "encoding/json"
  "fmt"
  "github.com/google/uuid"
  "github.com/muesli/smartcrop"
  "github.com/muesli/smartcrop/nfnt"
  "github.com/nfnt/resize"
  "github.com/rs/zerolog/log"
  "github.com/rverst/bwof-backend/pkg/helper"
  "github.com/rverst/bwof-backend/pkg/models"
  "image"
  "image/jpeg"
  "image/png"
  "mime/multipart"
  "net/http"
  "os"
  "path"
  "path/filepath"
  "regexp"
  "strings"
  "time"
)

const (
  fbGraphUrl  = "https://graph.facebook.com/v8.0"
  o_embedPost = "/instagram_oembed?url=%s"
  //o_embedPic  = "/instagram_oembed?url=%s&maxwidth=1920&fields=thumbnail_url,author_name,provider_name,provider_url"
)

var (
  typeRegex      = regexp.MustCompile("(?i)multipart/form-data")
  mimeRegex      = regexp.MustCompile("(?i)^image/(png|jpe?g)$")
  jpegRegex      = regexp.MustCompile("(?i)^jpe?g$")
  plainTextRegex = regexp.MustCompile(`(?i)^[\w\d\s\r\n\täöüÄÖÜ!?"'()&#@=+,\-\[\]]*$`)
  instaUrlRegex  = regexp.MustCompile(`(?i)^(?P<url>https?://www.instagram.com/\w+/\w+)/.*$`)

  errInvalidPost = fmt.Errorf("invalid instagram post url")
)

type pictureResponse struct {
  Id            string          `json:"id"`
  Disabled      bool            `json:"disabled"`
  Type          int             `json:"type"`
  Title         string          `json:"title"`
  Text          string          `json:"text"`
  OrigUrl       string          `json:"orig_url"`
  CroppedUrl    string          `json:"cropped_url"`
  ThumbUrl      string          `json:"thumb_url"`
  ThumbCropUrl  string          `json:"thumb_crop_url"`
  Width         int             `json:"width"`
  Height        int             `json:"height"`
  UseCrop       bool            `json:"use_crop"`
  TopCrop       image.Rectangle `json:"top_crop"`
  CroppedBounds image.Rectangle `json:"cropped_bounds"`
  Created       time.Time       `json:"created"`
  Edited        time.Time       `json:"edited"`
}

func fromPicture(p models.Picture) pictureResponse {
  r := pictureResponse{
    Type:          1,
    Id:            p.Id.String(),
    Disabled:      p.Disabled,
    Title:         p.Content.Title,
    Text:          p.Content.Text,
    OrigUrl:       p.OriginalUrl,
    ThumbUrl:      p.ThumbnailUrl,
    CroppedUrl:    p.CroppedUrl,
    ThumbCropUrl:  p.ThumbCroppedUrl,
    UseCrop:       p.UseCropped,
    TopCrop:       p.TopCrop,
    CroppedBounds: p.CroppedBounds,
    Created:       p.Uploaded,
    Width:         p.OriginalBounds.Dx(),
    Height:        p.OriginalBounds.Dy(),
  }
  return r
}

func fromInsta(i models.Instagram) pictureResponse {
  r := pictureResponse{
    Id:       i.Id.String(),
    Type:     2,
    Disabled: i.Disabled,
    Title:    i.Data.Type,
    Text:     i.Data.HTML,
    ThumbUrl: i.ThumbnailUrl,
    Created:  i.Uploaded,
    Edited:   i.Edited,
  }
  return r
}

func uploadPicture(w http.ResponseWriter, r *http.Request) {

  if !typeRegex.MatchString(r.Header.Get("Content-Type")) {
    log.Error().Msg("wrong content-type")
    _, _ = helper.WriteError(w, http.StatusBadRequest, "request Content-Type isn't multipart/form-data")
    return
  }

  err := r.ParseMultipartForm(32 << 18)
  if err != nil {
    log.Error().Err(err).Msg("parse multipartForm failed")
    _, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
    return
  }

  file, handler, err := r.FormFile("uploadFile")
  if err != nil {
    log.Error().Err(err).Msg("get file")
    _, _ = helper.WriteError(w, http.StatusBadRequest, "can't find 'uploadFile'")
    return
  }
  defer helper.Close(file, "uploadFile")

  mime := handler.Header.Get("Content-Type")
  if !mimeRegex.MatchString(mime) {
    m := fmt.Sprintf("unsupported file, mime type was: %s", mime)
    log.Error().Msg(m)
    _, _ = helper.WriteError(w, http.StatusBadRequest, m)
    return
  }

  title := r.FormValue("text")
  text := r.FormValue("text")
  if !plainTextRegex.MatchString(title) ||
    !plainTextRegex.MatchString(text) {
    err := fmt.Errorf("only plain text allowed in title/text")
    log.Warn().Err(err).Msg("editPictureContent")
    _, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
    return
  }

  picture, err := savePicture(r, file, handler, title, text)
  if err != nil {
    log.Error().Err(err).Msg("savePicture")
    _, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
    return
  }
  _, _ = helper.WriteJson(w, http.StatusOK, fromPicture(*picture))
}

func savePicture(r *http.Request, file multipart.File, handler *multipart.FileHeader, title, text string) (*models.Picture, error) {

  img, format, err := image.Decode(file)
  if err != nil {
    return nil, err
  }
  _ = file.Close()

  ext := "png"
  if jpegRegex.MatchString(format) {
    ext = "jpg"
  }

  id := uuid.New()
  fileName := fmt.Sprintf("orig.%s", ext)
  thumbName := fmt.Sprintf("thumb.%s", ext)
  dir := path.Join(pictureDir, id.String())
  err = os.Mkdir(dir, 0770)
  if err != nil {
    return nil, err
  }

  f, err := os.OpenFile(path.Join(dir, fileName), os.O_WRONLY|os.O_CREATE, 0660)
  if err != nil {
    return nil, err
  }
  defer helper.Close(f, fileName)
  t, err := os.OpenFile(path.Join(dir, thumbName), os.O_WRONLY|os.O_CREATE, 0660)
  if err != nil {
    return nil, err
  }
  defer helper.Close(t, thumbName)

  thumb := resize.Thumbnail(helper.ThumbnailSize, helper.ThumbnailSize, img, resize.Lanczos3)

  w := float64(img.Bounds().Dx())
  h := w / helper.TargetRatio

  if int(h) > img.Bounds().Dy() {
    h = float64(img.Bounds().Dy())
    w = h * helper.TargetRatio
  }

  analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
  topCrop, _ := analyzer.FindBestCrop(img, int(w), int(h))

  var e1 error
  var e2 error
  if ext == "png" {
    e1 = png.Encode(f, img)
    e2 = png.Encode(t, thumb)
  } else if ext == "jpg" {
    o := &jpeg.Options{Quality: 100}
    e1 = jpeg.Encode(f, img, o)
    e2 = jpeg.Encode(t, thumb, o)
  }

  if e1 != nil {
    return nil, e1
  }
  if e2 != nil {
    return nil, e2
  }

  picture := &models.Picture{
    Id:               id,
    OriginalBounds:   img.Bounds(),
    TopCrop:          topCrop,
    OriginalPath:     fileName,
    ThumbnailPath:    thumbName,
    OriginalUrl:      fmt.Sprintf("/pictures/%s/%s", id.String(), filepath.Base(fileName)),
    ThumbnailUrl:     fmt.Sprintf("/pictures/%s/%s", id.String(), filepath.Base(thumbName)),
    UploadedFilename: handler.Filename,
    Uploaded:         time.Now(),
    Content: models.Content{
      Title: title,
      Text:  text,
    },
  }

  //todo: get user
  user := r.Context().Value("user")
  if user == nil {
    picture.Uploader = "anonymous"
  }

  if err := insertNewPicture(picture); err != nil {
    _ = os.RemoveAll(filepath.Dir(f.Name()))
    return nil, err
  }
  return picture, nil
}

func uploadInstagram(w http.ResponseWriter, r *http.Request) {

  err := r.ParseMultipartForm(32 << 18)
  if err != nil {
    log.Error().Err(err).Msg("parse multipartForm failed")
    _, _ = helper.WriteError(w, http.StatusBadRequest, err.Error())
    return
  }

  url := r.FormValue("url")
  if url == "" {
    log.Error().Msg("missing url")
    _, _ = helper.WriteError(w, http.StatusBadRequest, "missing url")
    return
  }

  match := instaUrlRegex.FindStringSubmatch(url)
  for i, name := range instaUrlRegex.SubexpNames() {
    if i != 0 && name == "url" && len(match) > i {
      url = match[i]
    }
  }

  post, err := saveInstagram(r, url)
  if err != nil {
    log.Error().Err(err).Msg("error saving post")
    s := http.StatusInternalServerError
    if err == errInvalidPost {
      s = http.StatusBadRequest
    }
    _, _ = helper.WriteError(w, s, err.Error())
    return
  }

  _, _ = helper.WriteJson(w, http.StatusOK, fromInsta(*post))
}

func saveInstagram(r *http.Request, url string) (*models.Instagram, error) {

  uri_post := fmt.Sprintf(o_embedPost, url)
  log.Info().Str("uri", uri_post).Msg("saveInstagram")

  c := http.DefaultClient

  req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", fbGraphUrl, uri_post), nil)
  if err != nil {
    return nil, err
  }
  req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", instaToken))

  r1, err := c.Do(req)
  if err != nil {
    return nil, err
  }
  defer helper.CloseRC(r1.Body, "r1")

  d1 := models.InstaData{}
  err = json.NewDecoder(r1.Body).Decode(&d1)
  if err != nil {
    return nil, err
  }

  if d1.ThumbnailUrl != "" {
    d1.ThumbnailUrl = strings.Replace(d1.ThumbnailUrl, "\\u0026", "&", -1)
  }

  r2, err := c.Get(d1.ThumbnailUrl)
  if err != nil {
    return nil, err
  }
  defer r2.Body.Close()
  thumb, format, err := image.Decode(r2.Body)
  if err != nil {
    return nil, err
  }
  ext := "png"
  if jpegRegex.MatchString(format) {
    ext = "jpg"
  }
  id := uuid.New()
  thumbName := fmt.Sprintf("thumb.%s", ext)
  dir := path.Join(instaPostDir, id.String())
  err = os.Mkdir(dir, 0770)
  if err != nil {
    return nil, err
  }
  t, err := os.OpenFile(path.Join(dir, thumbName), os.O_WRONLY|os.O_CREATE, 0660)
  if err != nil {
    return nil, err
  }
  defer helper.Close(t, thumbName)
  if ext == "png" {
    err = png.Encode(t, thumb)
  } else if ext == "jpg" {
    o := &jpeg.Options{Quality: 100}
    err = jpeg.Encode(t, thumb, o)
  }
  if err != nil {
    return nil, err
  }

  post := &models.Instagram{
    Id:              id,
    Edited:          time.Now(),
    Disabled:        false,
    PostUrl:         url,
    ThumbnailBounds: thumb.Bounds(),
    ThumbnailPath:   thumbName,
    ThumbnailUrl:    fmt.Sprintf("/instagram/%s/%s", id.String(), filepath.Base(thumbName)),
    Uploaded:        time.Now(),
    Data:            d1,
  }

  //todo: get user
  user := r.Context().Value("user")
  if user == nil {
    post.Uploader = "anonymous"
  }

  if err := insertNewInstagram(post); err != nil {
    _ = os.RemoveAll(filepath.Dir(t.Name()))
    return nil, err
  }

  return post, nil
}
