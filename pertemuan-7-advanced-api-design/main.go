package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tugas1-go/pertemuan-7-advanced-api-design/helper"

	"tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	"tugas1-go/pertemuan-7-advanced-api-design/app/service"
	"tugas1-go/pertemuan-7-advanced-api-design/config"
	"tugas1-go/pertemuan-7-advanced-api-design/database"
	"tugas1-go/pertemuan-7-advanced-api-design/route"
)

func main() {
	// 1. Muat konfigurasi dan logger.
	config.LoadEnv()
	logger := config.NewLogger()
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		logger.Error("konfigurasi JWT_SECRET tidak valid")
		os.Exit(1)
	}

	// 2. Hubungkan aplikasi ke PostgreSQL.
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error(
			"gagal terhubung ke database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Rakit dependency dari repository ke service.
	roleRepository := repository.NewRoleRepository(pool)
	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat konfigurasi permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("konfigurasi permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository, permissions)
	achievementRepository := repository.NewAchievementRepository(pool)
	achievementService := service.NewAchievementService(achievementRepository)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	userService := service.NewUserService(userRepository, permissions)

	// 4. Rakit aplikasi Fiber.
	app := config.NewApp(logger, route.Dependencies{
		Pool: pool, JWT: jwtManager, Permissions: permissions,
		StudentService: studentService, AchievementService: achievementService,
		AuthService: authService, UserService: userService,
	})

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error(
				"server berhenti",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}()

	logger.Info(
		"server berjalan",
		slog.String("port", port),
	)

	// 5. Tunggu Ctrl+C atau sinyal penghentian.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")

	// Beri waktu maksimal 10 detik agar request yang sedang
	// diproses dapat diselesaikan sebelum server benar-benar berhenti.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error(
			"gagal menutup server dengan rapi",
			slog.String("error", err.Error()),
		)
	}

	logger.Info("server berhenti dengan rapi")
}
