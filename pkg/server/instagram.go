package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/rverst/bwof-backend/pkg/helper"
	"github.com/rverst/bwof-backend/pkg/models"
	"goji.io/pat"
	"net/http"
	"os"
	"path"
	"sort"
)

func getInstagram(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}

	post, err := getDbInstagram(id)
	if err != nil {
		res := models.Response{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}
		_, _ = helper.WriteJson(w, res.Status, res)
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromInsta(*post))
}

func getInstagrams(w http.ResponseWriter, _ *http.Request) {
	posts, err := getDbInstagrams()
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusNotFound, "")
		return
	}

	if len(posts) == 0 {
		_, _ = helper.WriteError(w, http.StatusNotFound, "no pictures found in database")
		return
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Uploaded.After(posts[j].Uploaded)
	})

	list := make([]pictureResponse, len(posts))
	for i, p := range posts {
		list[i] = fromInsta(p)
	}

	_, _ = helper.WriteJson(w, http.StatusOK, list)
}

func disableInstagram(w http.ResponseWriter, r *http.Request) {
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
	log.Info().Str("id", id.String()).Interface("body", body).Msg("disableInstagram")

	post, err := getDbInstagram(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	post.Disabled = body.Disable
	err = updateInstagram(post)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = helper.WriteJson(w, http.StatusOK, fromInsta(*post))
}

func deleteInstagram(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(pat.Param(r, "id"))
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse id: %s", id)
		_, _ = helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id: %s", id))
		return
	}
	log.Info().Str("id", id.String()).Msg("deleteInstagram")

	post, err := getDbInstagram(id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = deleteDbInstagram(post.Id)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p := path.Join(instaPostDir, post.Id.String())
	err = os.RemoveAll(p)
	if err != nil {
		_, _ = helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
