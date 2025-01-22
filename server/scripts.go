package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Script struct {
	ID   int    `json:"id"`
	File string `json:"file"`
}

func getScripts() []Script {
	path := getScriptsPath()
	files, err := os.ReadDir(path)
	if err != nil {
		log.Panic(err)
	}

	// Try to read existing scripts.json
	scriptsPath := filepath.Join(path, "scripts.json")
	var scripts []Script
	maxID := 0

	data, err := os.ReadFile(scriptsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Panic(err)
		}
		// File doesn't exist, initialize empty scripts slice
		scripts = []Script{}
	} else {
		// Parse existing scripts.json
		if err := json.Unmarshal(data, &scripts); err != nil {
			log.Panic(err)
		}
		// Find highest ID
		for _, script := range scripts {
			if script.ID > maxID {
				maxID = script.ID
			}
		}
	}

	// Check each file in the scripts directory
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".ts" && ext != ".js" {
			continue
		}

		// Check if file exists in scripts
		found := false
		for _, script := range scripts {
			if file.Name() == script.File {
				found = true
				break
			}
		}

		// If not found, add new entry
		if !found {
			maxID++
			newScript := Script{
				File: file.Name(),
				ID:   maxID,
			}
			scripts = append(scripts, newScript)
		}
	}

	// Save updated scripts back to file
	updatedData, err := json.MarshalIndent(scripts, "", "    ")
	if err != nil {
		log.Panic(err)
	}

	if err := os.WriteFile(scriptsPath, updatedData, 0644); err != nil {
		log.Panic(err)
	}

	return scripts
}

func readScript(file string) (string, error) {
	path := getScriptsPath()
	script := filepath.Join(path, file)
	b, err := os.ReadFile(script)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func updateScript(script Script, id int, command string) error {
	path := getScriptsPath()
	json_path := filepath.Join(path, "scripts.json")
	script_path := filepath.Join(path, script.File)

	scripts := getScripts()
	if slices.Contains(scripts, script) {
		i := slices.Index(scripts, script)
		scripts[i] = Script{
			ID:   id,
			File: script.File,
		}

		bytes, err := json.MarshalIndent(scripts, "", "    ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(json_path, bytes, 0644); err != nil {
			return err
		}
	}

	return os.WriteFile(script_path, []byte(command), 0644)
}

func writeScript(id int, file, data string) error {
	scripts := getScripts()
	scripts = append(scripts, Script{ID: id, File: file})
	// write json file

	json, err := json.MarshalIndent(scripts, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(getScriptsPath(), "scripts.json"), json, 0644)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(getScriptsPath(), file), []byte(data), 0644)
}

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

func getScriptsPath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	path := filepath.Dir(exe)
	return filepath.Join(path, "scripts")
}
