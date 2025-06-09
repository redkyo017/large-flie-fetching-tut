package main

import (
	"fmt"
	"kart-challenge/internal/app"
	order "kart-challenge/internal/orders"
	product "kart-challenge/internal/products"
	promo "kart-challenge/internal/promos"
	"kart-challenge/pkg/config"
	"kart-challenge/pkg/middleware"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.LoadConfig()

	// PROMO CODE MODULE
	promoCodeService := promo.NewService(cfg.MaxFileSizeMB, cfg.Environment, cfg.LocalCouponDirPath)

	defer func() {
		if err := promoCodeService.Close(); err != nil {
			panic(err)
		}
	}()

	// --- Initial Data Loading ---
	if err := promoCodeService.LoadPromoCodesFromURLs(cfg.CouponFileURLs); err != nil {
		log.Fatalf("Fatal error during initial promo code loading: %v", err)
		// In production, consider non-fatal error here if server can function partially.
	}

	// PRODUCT MODULE
	// productService := product.NewInMemoryProductService()
	// Product Service (uses in-memory repository internally)
	productService := product.NewService(product.NewInMemoryProductRepository())
	// ORDER MODULE
	orderService := order.NewService(order.NewInMemoryOrderRepository(), productService, promoCodeService)

	fiberApp := fiber.New(fiber.Config{
		AppName: "Food Ordering API Server",
	})
	fiberApp.Use(middleware.NewRecoveryMiddleware())
	fiberApp.Use(logger.New())

	handlers := &app.Handlers{
		PromoHandler:   promo.NewHandler(promoCodeService),
		ProductHandler: product.NewHandler(productService),
		OrderHandler:   order.NewHandler(orderService),
	}

	app.RegisterAPIRoutes(fiberApp, handlers)

	fiberApp.Static("/public", "./public", fiber.Static{
		ByteRange: true,
		Browse:    true,
	})
	log.Printf("Serving static files from './public' at '/public'")
	log.Println("Access OpenAPI HTML at: http://localhost:8080/public/openapi.html")

	// --- Start Server ---
	// Start server in a goroutine so it doesn't block graceful shutdown
	go func() {
		listenAddr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("Server starting on %s", listenAddr)
		if err := fiberApp.Listen(listenAddr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	// Create a channel to listen for OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c // Block until a signal is received
	log.Println("Shutting down server gracefully...")

	// Attempt to gracefully shut down the Fiber app
	if err := fiberApp.Shutdown(); err != nil {
		// Log the error but don't use Fatalf, allowing the program to exit
		log.Printf("Fiber app shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
