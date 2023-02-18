package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kamilkamilc/jals/config"
	"github.com/kamilkamilc/jals/model"
	"github.com/kamilkamilc/jals/store"
)

//go:embed static/index.html
var index []byte

type Handler struct {
	Storage store.RedisStorage
}

// basic generator without collision check, to be replaced
func basicGenerator(length int, useEmoji bool) string {
	rand.Seed(time.Now().UnixNano())

	const characters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	generated := make([]byte, length)
	for i := range generated {
		generated[i] = characters[rand.Intn(len(characters))]
	}
	return string(generated)
}

func (h *Handler) ApiPostLink(w http.ResponseWriter, r *http.Request) {
	// temporary, no checking for errors
	decoder := json.NewDecoder(r.Body)

	type postData struct {
		OriginalLink string `json:"originalLink"`
	}
	var data postData
	decoder.Decode(&data)
	shortLink := basicGenerator(8, false)
	h.Storage.SaveLink(&model.Link{
		ShortLink: shortLink,
		LinkInfo: model.LinkInfo{
			OriginalLink: data.OriginalLink,
			Clicks:       0,
		},
	})
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{\"shortLink\":\"%v\"}", shortLink))
}

func (h *Handler) ApiGetShortLink(w http.ResponseWriter, r *http.Request) {
	shortLink := chi.URLParam(r, "shortLink")
	w.Header().Set("Content-Type", "application/json")
	linkInfo, err := h.Storage.RetrieveLinkInfo(shortLink)
	if err != nil || linkInfo.OriginalLink == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
	} else {
		io.WriteString(w, fmt.Sprintf("{\"originalLink\":\"%v\",\"clicks\":\"%v\"}",
			linkInfo.OriginalLink, linkInfo.Clicks,
		))
	}
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(index)
}

func (h *Handler) GetShortLink(w http.ResponseWriter, r *http.Request) {
	shortLink := chi.URLParam(r, "shortLink")
	originalLink, err := h.Storage.RetrieveOriginalLink(shortLink)
	if err != nil || originalLink == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(index)
	} else {
		h.Storage.IncrementClicks(shortLink)
		http.Redirect(w, r, originalLink, http.StatusFound)
	}
}

func (h *Handler) PostLink(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// temporary, no checking for errors
	//useEmoji := r.Form["emoji"][0] == "on"
	originalLink := r.Form["link"][0]
	shortLink := basicGenerator(8, false)
	h.Storage.SaveLink(&model.Link{
		ShortLink: shortLink,
		LinkInfo: model.LinkInfo{
			OriginalLink: originalLink,
			Clicks:       0,
		},
	})
	w.Write([]byte(shortLink))
}

func GetHealthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func main() {
	appConfig := config.AppConfig()

	redisStorage := store.InitializeRedisStorage(appConfig)
	handler := &Handler{Storage: *redisStorage}

	apiRouter := chi.NewRouter()
	apiRouter.Post("/link", handler.ApiPostLink)
	apiRouter.Get("/link/{shortLink}", handler.ApiGetShortLink)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Mount("/api", apiRouter)

	r.Get("/", GetIndex)
	r.Get("/{shortLink}", handler.GetShortLink)
	r.Post("/link", handler.PostLink)

	r.Get("/healthz", GetHealthz)

	http.ListenAndServe(appConfig.Address, r)
}
