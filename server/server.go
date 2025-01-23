package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	fiberApp *fiber.App
	// Channel to signal when server is ready
	serverReady chan bool
)

// StartServer initializes and starts the Fiber server
func StartServer() {
	// Create the channel if it doesn't exist
	if serverReady == nil {
		serverReady = make(chan bool, 1)
	}

	go func() {
		if fiberApp != nil {
			fiberApp.Shutdown()
		}
		fiberApp = fiber.New()

		getScripts()

		fiberApp.Use(cors.New())

		fiberApp.Get("/scripts", fiberGetScripts)
		fiberApp.Get("/scripts/:id", executeScript)

		port := fyne.CurrentApp().Preferences().StringWithFallback("port", "9212")

		// Signal that the server is ready before starting to listen
		serverReady <- true

		fiberApp.Listen(":" + port)
	}()
}

// executeScript runs the specified script using bun
func executeScript(c *fiber.Ctx) error {
	path := getScriptsPath()
	id := c.Params("id")
	script, err := url.PathUnescape(filepath.Join(path, id))
	if err != nil {
		return err
	}

	proc := exec.Command("bun", "run", script)

	output, err := proc.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return c.SendString(strings.TrimSpace(string(output)))
}

// fiberGetScripts returns a list of available scripts
func fiberGetScripts(c *fiber.Ctx) error {
	scripts := getScripts()
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].ID < scripts[j].ID
	})
	for i, v := range scripts {
		scripts[i].File = strings.Replace(v.File, filepath.Ext(v.File), "", -1)
	}
	var out []string
	for _, v := range scripts {
		out = append(out, v.File)
	}
	return c.JSON(out)
}
