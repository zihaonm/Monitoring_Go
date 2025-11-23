package main

import (
	"fmt"
	"monitoring/handlers"
	"monitoring/models"
	"monitoring/services"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize persistence
	dataFile := "monitoring_data.json"
	persistence := models.NewPersistenceManager(dataFile)

	// Load existing data
	appData, err := persistence.Load()
	if err != nil {
		fmt.Printf("Warning: Could not load existing data: %v\n", err)
		appData = &models.AppData{
			Services:   make(map[string]*models.MonitoredService),
			Histories:  make(map[string]*models.ServiceHistory),
			TelegramConfig: &models.TelegramConfig{
				Enabled: false,
			},
		}
	}

	// Initialize store and load data
	store := models.NewServiceStore()
	store.LoadFromMap(appData.Services)

	// Initialize history store and load data
	historyStore := models.NewHistoryStore(100) // Keep last 100 checks per service
	historyStore.LoadFromMap(appData.Histories)

	// Initialize Telegram service and load config
	telegram := services.NewTelegramService()
	telegram.LoadConfig(appData.TelegramConfig)

	// Initialize system service early
	systemService := services.NewSystemService(telegram)
	systemService.LoadAlertConfig(appData.SystemAlertConfig)

	// Setup auto-save callback
	saveData := func() {
		data := &models.AppData{
			Services:          store.GetAllAsMap(),
			Histories:         historyStore.GetAllAsMap(),
			TelegramConfig:    telegram.GetRawConfig(),
			SystemAlertConfig: systemService.GetAlertConfig(),
		}
		if err := persistence.Save(data); err != nil {
			fmt.Printf("Error saving data: %v\n", err)
		}
	}

	// Set persistence callbacks
	store.SetPersistence(persistence, saveData)
	telegram.SetOnSave(saveData)

	// Initialize monitor service
	monitor := services.NewMonitorService(store, historyStore, telegram)

	// Initialize scheduler
	scheduler := services.NewScheduler(monitor)
	scheduler.Start()
	defer scheduler.Stop()

	// Initialize handlers
	serviceHandler := handlers.NewServiceHandler(store, historyStore, monitor)
	telegramHandler := handlers.NewTelegramHandler(telegram)
	systemHandler := handlers.NewSystemHandler(systemService)

	fmt.Printf("💾 Data will be saved to: %s\n", dataFile)
	if len(appData.Services) > 0 {
		fmt.Printf("📥 Loaded %d service(s) from disk\n", len(appData.Services))
	}

	// Setup Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Serve dashboard
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// API routes
	api := router.Group("/api")
	{
		// Service endpoints
		api.GET("/services", serviceHandler.GetAllServices)
		api.GET("/services/:id", serviceHandler.GetService)
		api.POST("/services", serviceHandler.CreateService)
		api.PUT("/services/:id", serviceHandler.UpdateService)
		api.DELETE("/services/:id", serviceHandler.DeleteService)
		api.POST("/services/:id/check", serviceHandler.CheckServiceNow)
		api.GET("/services/:id/statistics", serviceHandler.GetServiceStatistics)
		api.GET("/services/:id/history", serviceHandler.GetServiceHistory)

		// Telegram endpoints
		api.GET("/telegram/config", telegramHandler.GetConfig)
		api.PUT("/telegram/config", telegramHandler.UpdateConfig)
		api.POST("/telegram/test", telegramHandler.TestNotification)

		// System endpoints
		api.GET("/system/info", systemHandler.GetSystemInfo)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nShutting down server...")
		scheduler.Stop()
		os.Exit(0)
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\n🚀 Monitoring server started on http://localhost:%s\n", port)
	fmt.Printf("📊 Dashboard: http://localhost:%s\n", port)
	fmt.Printf("🔧 API: http://localhost:%s/api/services\n\n", port)

	if err := router.Run(":" + port); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
