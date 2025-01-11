package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getScripts() []string {
	exe, err := os.Executable()
	if err != nil {
		log.Panic(err)
	}
	path := filepath.Dir(exe)
	files, err := os.ReadDir(filepath.Join(path, "/scripts"))
	if err != nil {
		log.Panic(err)
	}
	tasks := []string{}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".ts" || ext == ".js" {
			tasks = append(tasks, file.Name())
		}
	}
	return tasks
}

func readScript(file string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	path := filepath.Dir(exe)
	script := filepath.Join(path, "scripts", file)
	b, err := os.ReadFile(script)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeScript(file, data string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	path := filepath.Dir(exe)
	script := filepath.Join(path, "scripts", file)

	return os.WriteFile(script, []byte(data), 0644)
}

func executeScript(c *fiber.Ctx) error {
	id := c.Params("id")

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	path := filepath.Dir(exe)
	script, err := url.PathUnescape(filepath.Join(path, "scripts", id))
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
