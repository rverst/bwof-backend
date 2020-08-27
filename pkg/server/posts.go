package server

import (
	"github.com/rs/zerolog/log"
	"github.com/rverst/bwof-backend/pkg/helper"
	"math/rand"
	"net/http"
  "sort"
  "time"
)

type item struct {
	Type   int    `json:"type"`
	Url    string `json:"url"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func getList(w http.ResponseWriter, _ *http.Request) {

	pics, err := getDbPictures()
	if err != nil {
		log.Error().Err(err).Msg("error fetching picture posts")
	}

	inst, err := getDbInstagrams()
	if err != nil {
		log.Error().Err(err).Msg("error fetching instagram posts")
	}

	list := make([]item, 0)

	if pics != nil {
		for _, p := range pics {
			if p.Disabled {
				continue
			}
			x := item{
				Type:  1,
				Title: p.Content.Title,
				Text:  p.Content.Text,
			}
			if p.UseCropped {
				x.Url = p.CroppedUrl
				x.Width = p.CroppedBounds.Dx()
				x.Height = p.CroppedBounds.Dy()
			} else {
				x.Url = p.OriginalUrl
				x.Width = p.OriginalBounds.Dx()
				x.Height = p.OriginalBounds.Dy()
			}

			list = append(list, x)
		}
	}

	if inst != nil {
		for _, i := range inst {
			if i.Disabled {
				continue
			}
			x := item{
				Type: 2,
				Title: i.Data.Type,
				Text: i.Data.HTML,
			}

			list = append(list, x)
		}
	}

	if len(list) == 0 {
		_, _ = helper.WriteError(w, http.StatusNoContent, "unable to fetch posts")
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	_, _ = helper.WriteJson(w, http.StatusOK, list)
}

func getPosts(w http.ResponseWriter, r *http.Request)  {
  pics, err1 := getDbPictures()
  inst, err2 := getDbInstagrams()
  if err1 != nil && err2 != nil {
    _, _ = helper.WriteError(w, http.StatusNotFound, "")
    return
  }

  if len(pics) == 0 && len(inst) == 0 {
    _, _ = helper.WriteError(w, http.StatusNotFound, "no posts found in database")
    return
  }

  list := make([]pictureResponse, 0)
  for _, p := range pics {
    list = append(list, fromPicture(p))
  }
  for _, p := range inst {
    list = append(list, fromInsta(p))
  }

  sort.Slice(list, func(i, j int) bool {
    return list[i].Created.After(list[j].Created)
  })




  _, _ = helper.WriteJson(w, http.StatusOK, list)
}
