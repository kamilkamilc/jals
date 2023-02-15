package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kamilkamilc/jals/config"
)

//go:embed static/index.html
var index []byte

// dummy in-memory storage
var linkStorage map[string]string

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

func ApiPostLink(w http.ResponseWriter, r *http.Request) {
	// temporary, no checking for errors
	decoder := json.NewDecoder(r.Body)

	type postData struct {
		OriginalLink string `json:"originalLink"`
	}
	var data postData
	decoder.Decode(&data)
	shortLink := basicGenerator(8, false)
	linkStorage[shortLink] = data.OriginalLink
	fmt.Printf("%+v\n%+v\n", data, shortLink)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"shortLink\":\"" + shortLink + "\"}"))
}

func ApiGetShortLink(w http.ResponseWriter, r *http.Request) {
	shortLink := chi.URLParam(r, "shortLink")
	w.Header().Set("Content-Type", "application/json")
	if originalLink, ok := linkStorage[shortLink]; ok {
		w.Write([]byte("{\"originalLink\":\"" + originalLink + "\"}"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
	}
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	w.Write(index)
}

func GetShortLink(w http.ResponseWriter, r *http.Request) {
	shortLink := chi.URLParam(r, "shortLink")
	if originalLink, ok := linkStorage[shortLink]; ok {
		http.Redirect(w, r, originalLink, http.StatusFound)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write(index)
	}
}

func PostLink(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	useEmoji := r.Form["emoji"][0] == "on"
	originalLink := r.Form["link"][0]
	shortLink := basicGenerator(8, useEmoji)
	fmt.Printf("%+v\n%+v\n%+v\n", useEmoji, originalLink, shortLink)
	linkStorage[shortLink] = originalLink
	w.Write(index)
}

func GetHealthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func main() {
	appConfig := config.AppConfig()

	linkStorage = make(map[string]string)

	apiRouter := chi.NewRouter()
	apiRouter.Post("/link", ApiPostLink)
	apiRouter.Get("/link/{shortLink}", ApiGetShortLink)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Mount("/api", apiRouter)

	r.Get("/", GetIndex)
	r.Get("/{shortLink}", GetShortLink)
	r.Post("/link", PostLink)

	r.Get("/healthz", GetHealthz)

	http.ListenAndServe(appConfig.Address, r)
}
