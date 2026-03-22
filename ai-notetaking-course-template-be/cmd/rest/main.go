package main

import (
	"ai-notetaking-be/internal/controller"
	"ai-notetaking-be/internal/helpers"
	"ai-notetaking-be/internal/loggers"
	"ai-notetaking-be/internal/middleware"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/internal/service"
	"ai-notetaking-be/pkg/database"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	apmFiber "go.elastic.co/apm/module/apmfiber"
)

func main() {
	log.Println("[MAIN] Application Started")
	helpers.SetupLogger()
	SetupLogfile()
	godotenv.Load()
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})
	logger := loggers.Logger{}
	app.Use(middleware.LoggingMiddleware(logger))
	app.Use(apmFiber.Middleware())

	app.Use(serverutils.ErrorHandlerMiddleware())

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))

	exampleRepository := repository.NewExampleRepository(db)
	notebookRepository := repository.NewNotebookRepository(db, logger)
	exampleService := service.NewExampleService(exampleRepository)
	notebookService := service.NewNotebookService(notebookRepository, db)

	exampleController := controller.NewExampleController(exampleService)
	notebookController := controller.NewNotebookController(notebookService)

	api := app.Group("/api")
	exampleController.RegisterRoutes(api)
	notebookController.RegisterRoutes(api)

	log.Fatal(app.Listen(":3000"))
}

func SetupLogfile() {
	logFile, err := os.OpenFile("./logs/log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}
