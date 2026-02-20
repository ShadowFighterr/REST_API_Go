package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	_, err = sqlconnect.ConnectDB("school_management")
	if err != nil {
		panic(err)
	}
	// defer db.Close()
	port := os.Getenv("API_PORT")
	cert := "cert.pem"
	key := "key.pem"

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
	secureMux := utils.ApplyMiddlewares(router.Router(),
		mw.Cors,
		rl.Limit,
		mw.ResponseTimeMiddleware,
		mw.CompressionMiddleware,
		mw.HPPMiddleware(hppOptions),
		mw.SecurityHeaders,
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
