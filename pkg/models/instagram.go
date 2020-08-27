package models

import (
  "github.com/google/uuid"
  "image"
  "time"
)

type InstaData struct {
  Version         string `json:"version"`
  AuthorName      string `json:"author_name"`
  ProviderName    string `json:"provider_name"`
  ProviderURL     string `json:"provider_url"`
  Type            string `json:"type"`
  Width           int    `json:"width"`
  HTML            string `json:"html"`
  ThumbnailUrl    string `json:"thumbnail_url"`
  ThumbnailWidth  int    `json:"thumbnail_width"`
  ThumbnailHeight int    `json:"thumbnail_height"`
}

type Instagram struct {
  Id              uuid.UUID       `json:"id"`
  Edited          time.Time       `json:"edited"`
  Disabled        bool            `json:"disabled"`
  PostUrl         string          `json:"post_url"`
  ThumbnailBounds image.Rectangle `json:"thumbnail_bounds"`
  ThumbnailPath   string          `json:"thumbnail_path"`
  ThumbnailUrl    string          `json:"thumbnail_url"`
  Uploaded        time.Time       `json:"uploaded"`
  Uploader        string          `json:"uploader"`
  Data            InstaData       `json:"data"`
}
