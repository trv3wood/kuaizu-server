package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/db"
	"github.com/trv3wood/kuaizu-server/internal/handler"
	"github.com/trv3wood/kuaizu-server/internal/middleware"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Printf("Starting Kuaizu Server %s (Commit: %s, Built at: %s)\n", version, commit, date)
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables\n")
	}

	// Initialize Echo
	e := echo.New()
	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Initialize database connection
	ctx := context.Background()
	pool, err := db.New(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to database")

	// Initialize repository and handler
	repo := repository.New(pool)
	server := handler.NewServer(repo)

	// Register API routes with /api/v2 prefix
	apiGroup := e.Group("/api/v2")

	// Add JWT authentication middleware with skipper for public endpoints
	jwtConfig := middleware.DefaultJWTConfig()
	jwtConfig.Skipper = func(c echo.Context) bool {
		path := c.Path()
		method := c.Request().Method

		// Public endpoints that don't require authentication
		publicEndpoints := []string{
			"/api/v2/auth/login/wechat",    // WeChat login
			"/api/v2/dictionaries/schools", // School list
			"/api/v2/dictionaries/majors",  // Major list
			"/api/v2/products",             // Product list
		}

		// Check exact matches
		for _, endpoint := range publicEndpoints {
			if path == endpoint {
				return true
			}
		}

		// Public GET endpoints with path parameters
		if method == "GET" {
			// /api/v2/projects - list (public)
			if path == "/api/v2/projects" {
				return true
			}
			// /api/v2/projects/:id - detail (public), but NOT /projects/my
			if strings.HasPrefix(path, "/api/v2/projects/") &&
				!strings.Contains(path, "/applications") &&
				path != "/api/v2/projects/my" {
				return true
			}
			// /api/v2/talent-profiles - list (public)
			if path == "/api/v2/talent-profiles" {
				return true
			}
			// /api/v2/talent-profiles/:id - detail (public)
			if strings.HasPrefix(path, "/api/v2/talent-profiles/") {
				return true
			}
		}

		return false
	}
	apiGroup.Use(middleware.JWTAuth(jwtConfig))

	api.RegisterHandlers(apiGroup, server)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}
