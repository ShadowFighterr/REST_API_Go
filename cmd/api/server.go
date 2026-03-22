package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbed() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading embedded .env file: %v", err)
	}
	tempfile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temporary file for .env: %v", err)
	}
	defer os.Remove(tempfile.Name())

	if _, err := tempfile.Write(content); err != nil {
		log.Fatalf("Error writing to temporary .env file: %v", err)
	}
	if err := tempfile.Close(); err != nil {
		log.Fatalf("Error closing temporary .env file: %v", err)
	}

	err = godotenv.Load(tempfile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file from embedded content: %v", err)
	}
}

func main() {

	loadEnvFromEmbed()

	port := os.Getenv("API_PORT")
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(100, 1*time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery:                   true,
		CheckBody:                    true,
		CheckBodyOnlyForContentTypes: "application/x-www-form-urlencoded",
		Whitelist:                    []string{"SortBy", "Asc", "name", "age", "class"},
	}

	jwtExcludedPaths := []string{
		"/execs/login",
		"/execs/forgotpassword",
		"/execs/resetpassword/reset/",
		"/execs/resetpassword/reset/*",
	}

	secureMux := utils.ApplyMiddlewares(
		router.MainRouter(),
		mw.SecurityHeaders,
		mw.CompressionMiddleware,
		mw.HPPMiddleware(hppOptions),
		mw.XSSProtectionMiddleware,
		mw.MiddlewaresExcludeRoutes(mw.JWTMiddleware, jwtExcludedPaths),
		mw.ResponseTimeMiddleware,
		rl.Limit,
		mw.Cors,
	)

	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("Server is running on port %s\n", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
