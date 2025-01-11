package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

		fiberApp.Get("/scripts", getListCustomTasks)
		fiberApp.Get("/scripts/:id", getExecuteCustomTask)

		port := preferences.StringWithFallback("port", "9212")

		fiberApp.Listen(":" + port)
	}()
}

func getExecuteCustomTask(c *fiber.Ctx) error {
	id := c.Params("id")

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	bun, err := exec.LookPath("bun")
	if err != nil {
		return err
	}

	path := filepath.Dir(exe)
	script := filepath.Join(path, "scripts", id)

	proc := exec.Command(bun, "run", script)

	output, err := proc.Output()
	if err != nil {
		return err
	}
	return c.SendString(strings.TrimSpace(string(output)))
}

func getListCustomTasks(c *fiber.Ctx) error {
	return c.JSON(getScripts())
}
