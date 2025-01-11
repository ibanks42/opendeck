package main

import (
	"log"
	"os"
	"path/filepath"
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
		tasks = append(tasks, file.Name())
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
