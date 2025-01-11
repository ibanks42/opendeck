package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var fiberApp *fiber.App

func StartServer() {
	go func() {
		if fiberApp != nil {
			fiberApp.Shutdown()
		} else {
			fiberApp = fiber.New()
		}

		getScripts()

		fiberApp.Use(cors.New())

		fiberApp.Get("/scripts", func(c *fiber.Ctx) error {
			return c.JSON(getScripts())
		})
		fiberApp.Get("/scripts/:id", executeScript)

		port := preferences.StringWithFallback("port", "9212")

		fiberApp.Listen(":" + port)
	}()
}
