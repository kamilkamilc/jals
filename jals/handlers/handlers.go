package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/kamilkamilc/jals/generator"
	"github.com/kamilkamilc/jals/model"
	"github.com/kamilkamilc/jals/store"
	"github.com/kamilkamilc/jals/views"
)

type Handler struct {
	Storage        store.Storage
	SessionStorage sessions.Store
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

func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	session, _ := h.SessionStorage.Get(r, "fmsgs")
	t := template.Must(template.New("index").Parse(string(views.IndexTemplate)))
	type ViewData struct {
		Flashes []interface{}
	}
	vd := ViewData{Flashes: session.Flashes("messages")}
	session.Save(r, w)
	t.Execute(w, vd)
}

func (h *Handler) GetShortLink(w http.ResponseWriter, r *http.Request) {
	shortLink := chi.URLParam(r, "shortLink")
	originalLink, err := h.Storage.RetrieveOriginalLink(shortLink)
	if err != nil || originalLink == "" {
		w.WriteHeader(http.StatusNotFound)
		session, _ := h.SessionStorage.Get(r, "fmsgs")
		session.AddFlash("Link not found", "warnings")
		t := template.Must(template.New("index").Parse(string(views.IndexTemplate)))
		type ViewData struct {
			Flashes []interface{}
		}
		vd := ViewData{Flashes: session.Flashes("warnings")}
		session.Save(r, w)
		t.Execute(w, vd)
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
	session, _ := h.SessionStorage.Get(r, "fmsgs")
	session.AddFlash("Short link: "+shortLink, "messages")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
