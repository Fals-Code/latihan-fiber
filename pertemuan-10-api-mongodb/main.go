package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-10-api-mongodb/app/repository"
	"tugas1-go/pertemuan-10-api-mongodb/app/service"
	"tugas1-go/pertemuan-10-api-mongodb/config"
	"tugas1-go/pertemuan-10-api-mongodb/database"
	portfolioRoute "tugas1-go/pertemuan-10-api-mongodb/route"
	p7repository "tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	p7service "tugas1-go/pertemuan-7-advanced-api-design/app/service"
	p7config "tugas1-go/pertemuan-7-advanced-api-design/config"
	p7database "tugas1-go/pertemuan-7-advanced-api-design/database"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
	p7route "tugas1-go/pertemuan-7-advanced-api-design/route"
)

func main() {
	p7config.LoadEnv()
	logger := p7config.NewLogger()
	if os.Getenv("DB_NAME") == "" {
		if err := os.Setenv("DB_NAME", config.PostgreSQLTestDatabase); err != nil {
			logger.Error("tidak dapat menetapkan database PostgreSQL Pertemuan 10")
			os.Exit(1)
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		logger.Error("konfigurasi JWT_SECRET tidak valid")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := p7database.NewPool(ctx)
	if err != nil {
		logger.Error("gagal terhubung ke database PostgreSQL")
		os.Exit(1)
	}
	defer pool.Close()

	roles, err := p7repository.NewRoleRepository(pool).LoadPermissions(ctx)
	if err != nil {
		logger.Error("gagal memuat permission dari PostgreSQL")
		os.Exit(1)
	}
	permissions := p7helper.NewPermissionSet(roles)
	studentRepository := p7repository.NewStudentRepository(pool)
	studentService := p7service.NewStudentService(studentRepository, permissions)
	achievementService := p7service.NewAchievementService(p7repository.NewAchievementRepository(pool))
	userRepository := p7repository.NewUserRepository(pool)
	jwtManager := p7helper.NewJWTManager(
		jwtSecret,
		p7config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(p7config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	authService := p7service.NewAuthService(
		userRepository,
		p7repository.NewTokenRepository(pool),
		jwtManager,
		permissions,
		time.Duration(p7config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	userService := p7service.NewUserService(userRepository, permissions)
	mongoClient, mongoErr := database.NewMongoClient(ctx, config.MongoURI())
	if mongoErr != nil {
		logger.Error("gagal terhubung ke MongoDB")
		pool.Close()
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Warn("gagal menutup koneksi MongoDB")
		}
	}()
	portfolioRepository := repository.NewPortfolioRepository(mongoClient.Database(config.MongoDatabaseName()))
	indexCtx, cancelIndexes := context.WithTimeout(ctx, 5*time.Second)
	if err := portfolioRepository.EnsureIndexes(indexCtx); err != nil {
		cancelIndexes()
		logger.Error("gagal menyiapkan index portfolio di MongoDB")
		_ = mongoClient.Disconnect(context.Background())
		pool.Close()
		os.Exit(1)
	}
	cancelIndexes()
	portfolioService := service.NewPortfolioService(portfolioRepository, studentRepository, permissions)
	app := p7config.NewApp(logger, p7route.Dependencies{
		Pool: pool, JWT: jwtManager, Permissions: permissions,
		StudentService: studentService, AchievementService: achievementService,
		AuthService: authService, UserService: userService,
	}, func(app *fiber.App) {
		portfolioRoute.RegisterPortfolioRoutes(app.Group("/api/v1"), portfolioService, jwtManager)
	})

	port := p7config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server Pertemuan 10 berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
}
