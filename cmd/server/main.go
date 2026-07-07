package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ezhigval/blog-cms-api/internal/auth"
	"github.com/ezhigval/blog-cms-api/internal/config"
	"github.com/ezhigval/blog-cms-api/internal/handler"
	"github.com/ezhigval/blog-cms-api/internal/middleware"
	"github.com/ezhigval/blog-cms-api/internal/repository"
	"github.com/ezhigval/blog-cms-api/internal/service"
	"github.com/ezhigval/go-toolkit/httputil"
	"github.com/ezhigval/go-toolkit/logger"
	tkmw "github.com/ezhigval/go-toolkit/middleware"
	tkpgx "github.com/ezhigval/go-toolkit/pgx"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(logger.Config{Level: cfg.LogLevel, Format: cfg.LogFormat})
	ctx := context.Background()

	pool, err := tkpgx.NewPool(ctx, tkpgx.Config{URL: cfg.DatabaseURL, MaxConns: 20})
	if err != nil {
		log.Error("postgres failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := tkpgx.Ping(ctx, pool); err != nil {
		log.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}
	_ = os.MkdirAll(cfg.UploadDir, 0o755)

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTTL)
	svc := service.NewCMS(cfg,
		repository.NewUserRepository(pool),
		repository.NewPostRepository(pool),
		repository.NewCategoryRepository(pool),
		repository.NewTagRepository(pool),
		repository.NewMediaRepository(pool),
		tokens,
	)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Use(tkmw.RequestID, tkmw.RealIP, tkmw.Recoverer(log), tkmw.AccessLog(log))
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	uploadsDir, _ := filepath.Abs(cfg.UploadDir)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	r.Get("/health", httputil.HealthHandler(map[string]func() error{
		"postgres": func() error { return tkpgx.Ping(ctx, pool) },
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Get("/posts", h.ListPostsPublic)
		r.Get("/posts/slug/{slug}", h.GetPostPublic)
		r.Get("/categories", h.ListCategories)
		r.Get("/tags", h.ListTags)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(tokens))
			r.Get("/auth/me", h.Me)
			r.Get("/admin/posts", h.ListPostsAdmin)
			r.Post("/admin/posts", h.CreatePost)
			r.Get("/admin/posts/{id}", h.GetPostAdmin)
			r.Put("/admin/posts/{id}", h.UpdatePost)
			r.Delete("/admin/posts/{id}", h.DeletePost)
			r.Post("/admin/categories", h.CreateCategory)
			r.Post("/admin/tags", h.CreateTag)
			r.Post("/admin/media", h.Upload)
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		log.Info("blog-cms-api started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
