package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kamilkamilc/jals/config"
	"github.com/kamilkamilc/jals/handlers"
	mid "github.com/kamilkamilc/jals/middleware"
	"github.com/kamilkamilc/jals/static"
	"github.com/kamilkamilc/jals/store"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(static.Index)
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
	handler := &handlers.Handler{Storage: redisStorage}

	apiRouter := chi.NewRouter()
	apiRouter.Post("/link", handler.ApiPostLink)
	apiRouter.Get("/link/{shortLink}", handler.ApiGetShortLink)

	r := chi.NewRouter()
	r.Use(mid.NewPrometheusMiddleware("jals"))
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Mount("/api", apiRouter)

	r.Get("/", GetIndex)
	r.Get("/{shortLink}", handler.GetShortLink)
	r.Post("/link", handler.PostLink)

	r.Get("/healthz", GetHealthz)
	r.Group(func(p chi.Router) {
		if appConfig.MetricsUser != "" && appConfig.MetricsPassword != "" {
			creds := make(map[string]string)
			creds[appConfig.MetricsUser] = appConfig.MetricsPassword
			p.Use(middleware.BasicAuth("Restricted", creds))
		}
		p.Method("GET", "/metrics", promhttp.Handler())
	})

	logger.Info().Str("address", appConfig.Address).Msg("server started")
	err := http.ListenAndServe(appConfig.Address, r)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
