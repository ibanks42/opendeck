package main

import (
	"path/filepath"
	"sort"
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

		fiberApp.Get("/scripts", func(c *fiber.Ctx) error {
			scripts := getScripts()
			sort.Slice(scripts, func(i, j int) bool {
				return scripts[i].ID < scripts[j].ID
			})
			for i, v := range scripts {
				scripts[i].File = strings.Replace(v.File, filepath.Ext(v.File), "", 0)
			}
			var out []string
			for _, v := range scripts {
				out = append(out, v.File)
			}
			return c.JSON(out)
		})
		fiberApp.Get("/scripts/:id", executeScript)

		port := preferences.StringWithFallback("port", "9212")

		fiberApp.Listen(":" + port)
	}()
}
