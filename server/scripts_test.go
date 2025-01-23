package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetScripts(t *testing.T) {
	testCases := []struct {
		name       string
		setupFiles []struct {
			name    string
			content string
		}
		setupJson   string
		wantErr     bool
		wantScripts int // Add explicit expected count
	}{
		{
			name: "normal case",
			setupFiles: []struct {
				name    string
				content string
			}{
				{
					name:    "test.js",
					content: "console.log('test')",
				},
			},
			setupJson:   `[{"id": 1, "file": "test.js"}]`,
			wantScripts: 1,
		},
		{
			name: "invalid json format",
			setupFiles: []struct {
				name    string
				content string
			}{
				{
					name:    "test.js",
					content: "console.log('test')",
				},
			},
			setupJson:   `{"invalid": "json"`,
			wantErr:     true,
			wantScripts: 1, // Should recover and find the file
		},
		{
			name: "empty but valid json",
			setupFiles: []struct {
				name    string
				content string
			}{
				{
					name:    "test.js",
					content: "console.log('test')",
				},
			},
			setupJson:   `[]`,
			wantScripts: 1, // Should find the file despite empty JSON
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test directory
			tmpDir := t.TempDir()
			originalPath := getScriptsPath
			getScriptsPath = func() string { return tmpDir }
			defer func() { getScriptsPath = originalPath }()

			// Create test files
			for _, f := range tc.setupFiles {
				err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Create scripts.json if content provided
			if tc.setupJson != "" {
				err := os.WriteFile(filepath.Join(tmpDir, "scripts.json"), []byte(tc.setupJson), 0644)
				if err != nil {
					t.Fatalf("Failed to create scripts.json: %v", err)
				}
			}

			// Run getScripts
			scripts := getScripts()

			// Verify results
			if len(scripts) != tc.wantScripts {
				t.Errorf("Expected %d scripts, got %d", tc.wantScripts, len(scripts))
			}

			// Verify scripts.json was created/updated correctly
			data, err := os.ReadFile(filepath.Join(tmpDir, "scripts.json"))
			if err != nil {
				t.Fatalf("Failed to read scripts.json: %v", err)
			}

			var savedScripts []Script
			if err := json.Unmarshal(data, &savedScripts); err != nil {
				t.Fatalf("Failed to parse saved scripts.json: %v", err)
			}

			if len(savedScripts) != tc.wantScripts {
				t.Errorf("Expected %d saved scripts, got %d", tc.wantScripts, len(savedScripts))
			}
		})
	}
}

func TestReadScript(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create test script
	testContent := "console.log('test')"
	err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading existing script
	content, err := readScript("test.js")
	if err != nil {
		t.Errorf("Failed to read script: %v", err)
	}
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}

	// Test reading non-existent script
	_, err = readScript("nonexistent.js")
	if err == nil {
		t.Error("Expected error when reading non-existent script")
	}
}

func TestUpdateScript(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create initial script
	script := Script{ID: 1, File: "test.js"}
	initialContent := "console.log('initial')"
	err := os.WriteFile(filepath.Join(tmpDir, script.File), []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create initial scripts.json
	initialScripts := []Script{script}
	scriptsData, _ := json.MarshalIndent(initialScripts, "", "\t") // Use tabs for indentation
	err = os.WriteFile(filepath.Join(tmpDir, "scripts.json"), scriptsData, 0644)
	if err != nil {
		t.Fatalf("Failed to create scripts.json: %v", err)
	}

	// Test updating script
	newContent := "console.log('updated')"
	err = updateScript(script, 1, newContent)
	if err != nil {
		t.Errorf("Failed to update script: %v", err)
	}

	// Verify content was updated
	content, err := readScript(script.File)
	if err != nil {
		t.Fatalf("Failed to read updated script: %v", err)
	}
	if content != newContent {
		t.Errorf("Expected content %q, got %q", newContent, content)
	}
}

func TestWriteScript(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Test writing new script
	testContent := "console.log('new script')"
	err := writeScript(1, "new.js", testContent)
	if err != nil {
		t.Errorf("Failed to write script: %v", err)
	}

	// Verify script was written
	content, err := readScript("new.js")
	if err != nil {
		t.Fatalf("Failed to read new script: %v", err)
	}
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}

	// Verify scripts.json was updated with correct formatting
	data, err := os.ReadFile(filepath.Join(tmpDir, "scripts.json"))
	if err != nil {
		t.Fatalf("Failed to read scripts.json: %v", err)
	}

	// Verify the JSON structure matches the expected format
	var scripts []Script
	if err := json.Unmarshal(data, &scripts); err != nil {
		t.Fatalf("Failed to parse scripts.json: %v", err)
	}

	// Check if the JSON is formatted with tabs
	expectedFormat := "[\n\t{\n\t\t\"id\": 1,\n\t\t\"file\": \"new.js\"\n\t}\n]"
	formattedData := string(data)
	if formattedData != expectedFormat {
		t.Errorf("Expected JSON format:\n%s\n\nGot:\n%s", expectedFormat, formattedData)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script in scripts.json, got %d", len(scripts))
	}
}

func TestGetMaxScriptId(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create test scripts
	scripts := []Script{
		{ID: 1, File: "test1.js"},
		{ID: 3, File: "test2.js"},
		{ID: 2, File: "test3.js"},
	}

	scriptsData, _ := json.MarshalIndent(scripts, "", "    ")
	err := os.WriteFile(filepath.Join(tmpDir, "scripts.json"), scriptsData, 0644)
	if err != nil {
		t.Fatalf("Failed to create scripts.json: %v", err)
	}

	maxId := getMaxScriptId()
	if maxId != 3 {
		t.Errorf("Expected max ID 3, got %d", maxId)
	}
}
