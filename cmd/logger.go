package cmd

import (
	"fmt"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func NewRequestLogger() echo.MiddlewareFunc {
	return echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			// Color codes
			const (
				reset   = "\033[0m"
				gray    = "\033[90m"
				cyan    = "\033[36m"
				blue    = "\033[34m"
				green   = "\033[32m"
				yellow  = "\033[33m"
				red     = "\033[31m"
				magenta = "\033[35m"
			)

			// Skip health checks
			if v.URI == "/health" {
				return nil
			}

			// Method color
			methodColor := cyan
			switch v.Method {
			case "GET":
				methodColor = blue
			case "POST":
				methodColor = green
			case "PUT":
				methodColor = yellow
			case "DELETE":
				methodColor = red
			case "PATCH":
				methodColor = magenta
			}

			// Status color
			statusColor := green
			if v.Status >= 500 {
				statusColor = red
			} else if v.Status >= 400 {
				statusColor = yellow
			} else if v.Status >= 300 {
				statusColor = cyan
			}

			// Format: timestamp method uri status latency
			fmt.Printf("%s%s%s %s%s%s %s %s%d%s %s%v%s\n",
				gray, v.StartTime.Format("2006/01/02 15:04:05"), reset,
				methodColor, v.Method, reset,
				v.URI,
				statusColor, v.Status, reset,
				gray, v.Latency, reset,
			)

			// Print detailed error information for error responses
			if v.Status >= 400 {
				// Print request details
				fmt.Printf("%s[ERROR DETAILS]%s\n", red, reset)
				fmt.Printf("  %sRequest:%s %s %s\n", yellow, reset, v.Method, v.URI)

				// Print request headers (excluding sensitive ones)
				if c.Request() != nil {
					fmt.Printf("  %sHeaders:%s\n", yellow, reset)
					for key, values := range c.Request().Header {
						// Skip sensitive headers
						if key == "Authorization" || key == "Cookie" {
							fmt.Printf("    %s: %s[REDACTED]%s\n", key, gray, reset)
							continue
						}
						for _, value := range values {
							fmt.Printf("    %s: %s\n", key, value)
						}
					}
				}

				// Print user info if available
				if userID := c.Get("user_id"); userID != nil {
					fmt.Printf("  %sUser ID:%s %v\n", yellow, reset, userID)
				}
				if openID := c.Get("open_id"); openID != nil {
					fmt.Printf("  %sOpen ID:%s %v\n", yellow, reset, openID)
				}

				// Print error if present
				if v.Error != nil {
					fmt.Printf("  %sError:%s %v\n", red, reset, v.Error)
				}
			}

			return nil
		},
	})
}
