package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/db"
)

type Server struct{}

// Ensure Server implements the generated ServerInterface
var _ api.ServerInterface = (*Server)(nil)

func (s *Server) GetHello(ctx echo.Context) error {
	return ctx.JSON(200, map[string]string{"message": "Hello World"})
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found, using environment variables\n")
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize database connection
	// We only log error for now as we might not have DB set up yet for simple run
	_, err := db.New(context.Background())
	if err != nil {
		fmt.Printf("Warning: Could not connect to database: %v\n", err)
	} else {
		fmt.Println("Connected to database")
	}

	server := &Server{}

	// Register generated handlers
	api.RegisterHandlers(e, server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(e.Start(":" + port))
}
