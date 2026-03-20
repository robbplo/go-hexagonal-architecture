package main

import (
	"context"
	"fmt"
	"os"

	"github.com/linkai/go-chatbot-api/adapters/logging"
	"github.com/linkai/go-chatbot-api/adapters/postgres"
	"github.com/linkai/go-chatbot-api/adapters/supabaseauth"
	"github.com/linkai/go-chatbot-api/adapters/system"
	"github.com/linkai/go-chatbot-api/app"
	"github.com/linkai/go-chatbot-api/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.ValidateBootstrapAdmin(); err != nil {
		fmt.Fprintf(os.Stderr, "validate bootstrap config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.New(cfg.Log.Level)

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

	authClient := supabaseauth.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.AuthUserPath, cfg.Supabase.AdminPath, cfg.AI.Timeout)
	userRepo := postgres.NewUserRepo(pool)
	clock := system.NewClock()

	service := app.NewBootstrapAdminService(userRepo, authClient, clock, logger)
	profile, err := service.EnsureAdmin(ctx, cfg.AdminBootstrap.Email, cfg.AdminBootstrap.Password)
	if err != nil {
		logger.Error(ctx, "bootstrap admin", "error", err)
		os.Exit(1)
	}

	logger.Info(ctx, "bootstrap admin complete", "user_id", profile.ID, "email", profile.Email)
}
