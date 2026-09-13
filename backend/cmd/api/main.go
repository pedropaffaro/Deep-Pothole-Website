package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/pedropaffaro/deep-pothole-backend/internal/adapters/d1"
	"github.com/pedropaffaro/deep-pothole-backend/internal/adapters/onnx"
	"github.com/pedropaffaro/deep-pothole-backend/internal/adapters/s3"
	"github.com/pedropaffaro/deep-pothole-backend/internal/complaint"
	"github.com/pedropaffaro/deep-pothole-backend/internal/config"
	"github.com/pedropaffaro/deep-pothole-backend/internal/httpapi"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("encerrando por falha", "erro", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	repo := d1.New(cfg.D1AccountID, cfg.D1DatabaseID, cfg.D1APIToken)

	photos, err := s3.New(ctx, s3.Options{
		Bucket:        cfg.S3Bucket,
		Region:        config.S3Region,
		Endpoint:      cfg.S3Endpoint,
		AccessKey:     cfg.S3AccessKeyID,
		SecretKey:     cfg.S3SecretAccessKey,
		PublicBaseURL: cfg.S3PublicBaseURL,
	})
	if err != nil {
		return err
	}

	detector, err := onnx.New()
	if err != nil {
		return err
	}
	defer detector.Close()

	svc := complaint.NewService(repo, photos, detector)
	handler := httpapi.NewHandler(svc, config.RequestTimeout, log)

	app := fiber.New(fiber.Config{
		BodyLimit:             config.MaxUploadBytes,
		DisableStartupMessage: true,
		ReadTimeout:           config.RequestTimeout + 5*time.Second,
		WriteTimeout:          config.RequestTimeout + 5*time.Second,
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{AllowOrigins: config.CORSAllowedOrigins}))

	handler.Register(app.Group("/api/v1"))

	return listenAndServe(app, log)
}

func listenAndServe(app *fiber.App, log *slog.Logger) error {
	errCh := make(chan error, 1)

	go func() {
		log.Info("API no ar", "porta", config.Port)
		if err := app.Listen(":" + config.Port); err != nil {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("servidor HTTP: %w", err)
	case sig := <-stop:
		log.Info("sinal recebido, encerrando", "sinal", sig.String())
		return app.ShutdownWithTimeout(10 * time.Second)
	}
}
