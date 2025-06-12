package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"abt-analytics/internal/config"
	"abt-analytics/internal/controllers"
	"abt-analytics/internal/models"
	"abt-analytics/internal/repository"
	"abt-analytics/internal/services"
	_ "abt-analytics/docs" // Import generated docs
)

// @title ABT Analytics API
// @version 1.0
// @description Analytics dashboard API for ABT Corporation
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func main() {
	// Load configuration
	cfg := config.Load()

	// Try to connect to database (optional for demo)
	db, err := connectDatabase(cfg)
	var analyticsController *controllers.AnalyticsController
	
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		log.Println("Running in demo mode without database functionality")
		// Create a mock controller for demo purposes
		analyticsController = controllers.NewAnalyticsController(nil)
	} else {
		// Auto migrate
		if err := db.AutoMigrate(&models.Transaction{}); err != nil {
			log.Printf("Warning: Failed to migrate database: %v", err)
		}

		// Seed data
		if err := seedData(db); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		}

		// Initialize repository, services and controllers
		analyticsRepo := repository.NewAnalyticsRepository(db)
		analyticsService := services.NewAnalyticsService(analyticsRepo)
		analyticsController = controllers.NewAnalyticsController(analyticsService)
	}

	// Setup router
	router := setupRouter(analyticsController)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📚 Swagger documentation: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Printf("🏥 Health check: http://localhost:%s/api/v1/health", cfg.Port)
	log.Printf("📊 Analytics API: http://localhost:%s/api/v1/analytics/", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	
	// Try to connect to database with shorter timeout for demo
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	return db, err
}

func setupRouter(analyticsController *controllers.AnalyticsController) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://frontend", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", analyticsController.HealthCheck)
		
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/country-revenue", analyticsController.GetCountryRevenue)
			analytics.GET("/top-products", analyticsController.GetTopProducts)
			analytics.GET("/monthly-sales", analyticsController.GetMonthlySales)
			analytics.GET("/top-regions", analyticsController.GetTopRegions)
		}
	}

	return router
}

func seedData(db *gorm.DB) error {
	// Check if data already exists
	var count int64
	db.Model(&models.Transaction{}).Count(&count)
	if count > 0 {
		log.Println("Data already seeded, skipping...")
		return nil
	}

	log.Println("Seeding database with sample data...")
	
	// Create sample data seeder
	seeder := services.NewDataSeeder(db)
	return seeder.SeedData()
}
