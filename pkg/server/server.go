package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rs/zerolog/log"
	"github.com/rverst/bwof-backend/pkg/helper"
	goji "goji.io"
	"goji.io/pat"
	"net/http"
	"os"
	"path"
)

const (
	EnvDataDir              = "DATA"
	EnvDbFile               = "DB"
	EnvInstagramAccessToken = "INSTA_TOKEN"
)

var (
	db           *bolt.DB
	dataDir      string
	pictureDir   string
	instaPostDir string
	instaToken   string
)

func Run() {

	instaToken = os.Getenv(EnvInstagramAccessToken)

	var err error
	dataDir = helper.GetStringEnv(EnvDataDir, "/data")
	if len(dataDir) > 1 && dataDir[0] == '.' && dataDir[1] == '/' {
		dataDir = dataDir[2:]
	}

	pictureDir = path.Join(dataDir, "pictures")
	_, err = os.Stat(pictureDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(pictureDir, 0770)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open/create picture dir")
	}

	instaPostDir = path.Join(dataDir, "instagram")
	_, err = os.Stat(instaPostDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(instaPostDir, 0770)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open/create instagram dir")
	}

	dbFile := path.Join(dataDir, helper.GetStringEnv(EnvDbFile, "data.db"))
	db, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open db")
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Error().Err(err).Msg("db.Close()")
		}
	}()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Options("/*"), cors(blank))
	mux.HandleFunc(pat.Get("/api/list"), cors(getList))
  mux.HandleFunc(pat.Get("/api/posts"), cors(getPosts))

	mux.HandleFunc(pat.Get("/api/picture/:id"), cors(getPicture))
	mux.HandleFunc(pat.Get("/api/picture"), cors(getPictures))
	mux.HandleFunc(pat.Post("/api/picture"), cors(uploadPicture))
	mux.HandleFunc(pat.Patch("/api/picture/:id/crop"), cors(cropPicture))
	mux.HandleFunc(pat.Patch("/api/picture/:id/edit"), cors(editPictureContent))
	mux.HandleFunc(pat.Patch("/api/picture/:id/disable"), cors(disablePicture))
	mux.HandleFunc(pat.Delete("/api/picture/:id"), cors(deletePicture))

	mux.HandleFunc(pat.Get("/api/instagram/:id"), cors(getInstagram))
	mux.HandleFunc(pat.Get("/api/instagram"), cors(getInstagrams))
	mux.HandleFunc(pat.Post("/api/instagram"), cors(uploadInstagram))
	mux.HandleFunc(pat.Patch("/api/instagram/:id/disable"), cors(disableInstagram))
	mux.HandleFunc(pat.Delete("/api/instagram/:id"), cors(deleteInstagram))

	mux.Handle(pat.Get("/pictures/*"),
		http.StripPrefix("/pictures/", http.FileServer(http.Dir(pictureDir))))
	mux.Handle(pat.Get("/instagram/*"),
		http.StripPrefix("/instagram/", http.FileServer(http.Dir(instaPostDir))))
	mux.Handle(pat.Get("/*"), http.FileServer(http.Dir("/app/public")))

	err = http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start listener")
	}
	log.Info().Msg("service running at :8000")
}

func blank(w http.ResponseWriter, r *http.Request) {
	fmt.Println("blank", r.URL)
	w.WriteHeader(http.StatusOK)
}

func cors(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods",
			"OPTIONS, POST, GET, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		f(w, r) // original function call
	}
}
