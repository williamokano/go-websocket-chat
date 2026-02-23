package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/williamokano/example-websocket-chat/internal/auth"
	"github.com/williamokano/example-websocket-chat/internal/chat"
	"github.com/williamokano/example-websocket-chat/internal/config"
	"github.com/williamokano/example-websocket-chat/internal/user"
)

type Server struct {
	cfg    *config.Config
	router *chi.Mux
	hub    *chat.Hub
}

func New(cfg *config.Config, pool *pgxpool.Pool) (*Server, error) {
	userRepo := user.NewRepository(pool)
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	authService := auth.NewService(userRepo, jwtService)
	authHandler := auth.NewHandler(authService)
	authMiddleware := auth.NewMiddleware(jwtService)

	hub, err := chat.NewHub(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("creating chat hub: %w", err)
	}

	chatHandler := chat.NewHandler(hub, jwtService, cfg.WebSocketOriginPatterns())

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.With(authMiddleware.Authenticate).Get("/me", authHandler.Me)
	})

	r.Get("/ws", chatHandler.HandleWebSocket)

	return &Server{cfg: cfg, router: r, hub: hub}, nil
}

func (s *Server) Run(ctx context.Context) error {
	go s.hub.Run(ctx)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: s.router,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", s.cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
