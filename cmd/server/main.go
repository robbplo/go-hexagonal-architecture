package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linkai/go-chatbot-api/adapters/files"
	httpapi "github.com/linkai/go-chatbot-api/adapters/http"
	"github.com/linkai/go-chatbot-api/adapters/logging"
	"github.com/linkai/go-chatbot-api/adapters/openai"
	"github.com/linkai/go-chatbot-api/adapters/postgres"
	"github.com/linkai/go-chatbot-api/adapters/supabaseauth"
	"github.com/linkai/go-chatbot-api/adapters/supabasestorage"
	"github.com/linkai/go-chatbot-api/adapters/system"
	"github.com/linkai/go-chatbot-api/app"
	"github.com/linkai/go-chatbot-api/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.ValidateServer(); err != nil {
		fmt.Fprintf(os.Stderr, "validate server config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.Log.Level)

	shutdownTracer, err := initTracer(ctx, cfg)
	if err != nil {
		logger.Error(ctx, "initialize tracer", "error", err)
		os.Exit(1)
	}
	defer shutdownTracer()

	pool, err := postgres.OpenPool(ctx, cfg.Database.DSN)
	if err != nil {
		logger.Error(ctx, "open database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := postgres.RunMigrations(ctx, pool); err != nil {
		logger.Error(ctx, "run migrations", "error", err)
		os.Exit(1)
	}

	clock := system.NewClock()
	ids := system.NewIDGenerator()
	tokenCounter := system.NewApproxTokenCounter()

	companyRepo := postgres.NewCompanyRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	chatbotRepo := postgres.NewChatbotRepo(pool)
	grantRepo := postgres.NewGrantRepo(pool)
	conversationRepo := postgres.NewConversationRepo(pool)
	usageRepo := postgres.NewUsageRepo(pool)

	authClient := supabaseauth.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.AuthUserPath, cfg.Supabase.AdminPath, cfg.AI.Timeout)
	storageClient := supabasestorage.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.StorageBucket, cfg.AI.Timeout)
	extractor := files.NewExtractor()
	aiClient := openai.NewClient(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Timeout)

	companyService := app.NewCompanyService(companyRepo, userRepo, usageRepo, authClient, clock, ids, logger)
	userService := app.NewUserService(userRepo, companyRepo, authClient, clock, logger)
	sessionService := app.NewSessionService(userRepo, clock, logger)
	chatbotService := app.NewChatbotService(
		chatbotRepo,
		companyRepo,
		grantRepo,
		storageClient,
		extractor,
		tokenCounter,
		clock,
		ids,
		logger,
		cfg.Upload.MaxFileBytes,
		cfg.Upload.MaxFilesPerBot,
		cfg.AI.KnowledgeMaxTokens,
		cfg.Upload.AllowedTypes,
	)
	chatService := app.NewChatService(companyRepo, chatbotRepo, grantRepo, conversationRepo, usageRepo, aiClient, tokenCounter, clock, ids, logger, cfg.AI.Model, cfg.AI.HistoryMaxTokens)
	catalogService := app.NewCatalogService(chatbotRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	api := httpapi.NewAPI(mux)
	handler := httpapi.NewHandler(companyService, userService, chatbotService, chatService, catalogService, cfg.Supabase.InviteRedirectURL)
	handler.Register(api, httpapi.AuthMiddleware(api, authClient, sessionService), httpapi.AdminOnlyMiddleware(api))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error(context.Background(), "shutdown server", "error", err)
		}
	}()

	logger.Info(ctx, "server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(ctx, "server exited", "error", err)
		os.Exit(1)
	}
}

func initTracer(ctx context.Context, cfg config.Config) (func(), error) {
	if !cfg.Otel.Enabled {
		return func() {}, nil
	}

	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	resource, err := sdkresource.New(
		ctx,
		sdkresource.WithAttributes(semconv.ServiceName(cfg.Otel.ServiceName)),
	)
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = provider.Shutdown(shutdownCtx)
	}, nil
}
