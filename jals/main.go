package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"

	"github.com/kamilkamilc/jals/config"
	"github.com/kamilkamilc/jals/generator"
	"github.com/kamilkamilc/jals/model"
	"github.com/kamilkamilc/jals/store"
)

//go:embed static/index.html
var index []byte

type Handler struct {
	Storage store.Storage
}

func (h *Handler) ApiPostLink(w http.ResponseWriter, r *http.Request) {
	// temporary, no checking for errors
	decoder := json.NewDecoder(r.Body)

	type postData struct {
		OriginalLink string `json:"originalLink"`
	}
	var data postData
	decoder.Decode(&data)
	shortLink := generator.BasicGenerator(8, false)
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
	originalLink := r.Form["link"][0]
	shortLink := generator.BasicGenerator(8, false)
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
	logger := httplog.NewLogger("jals", httplog.Options{
		JSON: true,
	})

	redisStorage := store.InitializeRedisStorage(appConfig)
	handler := &Handler{Storage: redisStorage}

	apiRouter := chi.NewRouter()
	apiRouter.Post("/link", handler.ApiPostLink)
	apiRouter.Get("/link/{shortLink}", handler.ApiGetShortLink)

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Mount("/api", apiRouter)

	r.Get("/", GetIndex)
	r.Get("/{shortLink}", handler.GetShortLink)
	r.Post("/link", handler.PostLink)

	r.Get("/healthz", GetHealthz)

	logger.Info().Str("address", appConfig.Address).Msg("server started")
	err := http.ListenAndServe(appConfig.Address, r)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
