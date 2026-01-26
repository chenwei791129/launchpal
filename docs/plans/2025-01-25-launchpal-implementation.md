# LaunchPal Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a macOS menu bar application for managing LaunchAgents with a modern dark-themed GUI.

**Architecture:** Wails v2 desktop app with Go backend handling launchctl commands and plist operations, Nuxt 4 frontend providing Podman Desktop-style UI. Single `.app` bundle distribution.

**Tech Stack:** Go 1.21+, Wails v2, Nuxt 4, TypeScript, TailwindCSS

---

## Task 1: Initialize Wails Project

**Files:**
- Create: `wails.json`
- Create: `main.go`
- Create: `app.go`

**Step 1: Install Wails CLI (if not installed)**

Run:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Step 2: Verify Wails installation**

Run:
```bash
wails doctor
```
Expected: All checks pass

**Step 3: Initialize Wails project with vanilla template**

Run:
```bash
cd /Users/jeff/Desktop/macos-launchctl-controler/.worktrees/feature-launchpal
wails init -n launchpal -t vanilla
```
Expected: Project created in `launchpal/` directory

**Step 4: Move project files to root**

Run:
```bash
mv launchpal/* .
mv launchpal/.* . 2>/dev/null || true
rmdir launchpal
```

**Step 5: Verify project structure**

Run:
```bash
ls -la
```
Expected: See `wails.json`, `main.go`, `app.go`, `frontend/`

**Step 6: Test initial build**

Run:
```bash
wails build
```
Expected: Build succeeds, creates `build/bin/launchpal.app`

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: initialize Wails v2 project structure"
```

---

## Task 2: Setup Nuxt 4 Frontend

**Files:**
- Remove: `frontend/` (vanilla template)
- Create: `frontend/nuxt.config.ts`
- Create: `frontend/package.json`
- Create: `frontend/app.vue`

**Step 1: Remove vanilla frontend**

Run:
```bash
rm -rf frontend
```

**Step 2: Create Nuxt 4 project**

Run:
```bash
pnpm create nuxt frontend -t v4 --no-install --no-gitInit
```

**Step 3: Install frontend dependencies**

Run:
```bash
cd frontend && pnpm install
```

**Step 4: Configure Nuxt for Wails (static generation)**

Modify: `frontend/nuxt.config.ts`

```typescript
export default defineNuxtConfig({
  compatibilityDate: '2025-01-25',
  ssr: false,
  devtools: { enabled: false },
  app: {
    baseURL: './',
    buildAssetsDir: 'assets',
  },
  vite: {
    build: {
      assetsInlineLimit: 0,
    },
  },
})
```

**Step 5: Update wails.json for Nuxt**

Modify: `wails.json`

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "launchpal",
  "outputfilename": "launchpal",
  "frontend:install": "pnpm install",
  "frontend:build": "pnpm generate",
  "frontend:dev:watcher": "pnpm dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Jeff",
    "email": ""
  }
}
```

**Step 6: Add generate script to frontend/package.json**

Add to scripts section:
```json
"generate": "nuxt generate"
```

**Step 7: Test Wails dev mode**

Run:
```bash
cd /Users/jeff/Desktop/macos-launchctl-controler/.worktrees/feature-launchpal
wails dev
```
Expected: Window opens with Nuxt welcome page

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: replace vanilla frontend with Nuxt 4"
```

---

## Task 3: Install TailwindCSS and Setup Dark Theme

**Files:**
- Modify: `frontend/nuxt.config.ts`
- Create: `frontend/assets/css/main.css`
- Create: `frontend/tailwind.config.ts`

**Step 1: Install TailwindCSS module**

Run:
```bash
cd frontend && pnpm add -D @nuxtjs/tailwindcss
```

**Step 2: Add module to nuxt.config.ts**

```typescript
export default defineNuxtConfig({
  compatibilityDate: '2025-01-25',
  ssr: false,
  devtools: { enabled: false },
  modules: ['@nuxtjs/tailwindcss'],
  app: {
    baseURL: './',
    buildAssetsDir: 'assets',
  },
  vite: {
    build: {
      assetsInlineLimit: 0,
    },
  },
})
```

**Step 3: Create TailwindCSS config**

Create: `frontend/tailwind.config.ts`

```typescript
import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: [],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#a855f7',
          50: '#faf5ff',
          100: '#f3e8ff',
          200: '#e9d5ff',
          300: '#d8b4fe',
          400: '#c084fc',
          500: '#a855f7',
          600: '#9333ea',
          700: '#7e22ce',
          800: '#6b21a8',
          900: '#581c87',
        },
        surface: {
          DEFAULT: '#1e1e1e',
          50: '#3d3d3d',
          100: '#2d2d2d',
          200: '#252525',
          300: '#1e1e1e',
          400: '#171717',
          500: '#121212',
        },
      },
    },
  },
  plugins: [],
} satisfies Config
```

**Step 4: Create global CSS**

Create: `frontend/assets/css/main.css`

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  color-scheme: dark;
}

body {
  @apply bg-surface-400 text-gray-100;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Wails window drag region */
.wails-drag {
  --wails-draggable: drag;
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  @apply bg-surface-400;
}

::-webkit-scrollbar-thumb {
  @apply bg-surface-50 rounded;
}

::-webkit-scrollbar-thumb:hover {
  @apply bg-gray-500;
}
```

**Step 5: Reference CSS in nuxt.config.ts**

```typescript
export default defineNuxtConfig({
  compatibilityDate: '2025-01-25',
  ssr: false,
  devtools: { enabled: false },
  modules: ['@nuxtjs/tailwindcss'],
  css: ['~/assets/css/main.css'],
  app: {
    baseURL: './',
    buildAssetsDir: 'assets',
  },
  vite: {
    build: {
      assetsInlineLimit: 0,
    },
  },
})
```

**Step 6: Test dark theme**

Run:
```bash
wails dev
```
Expected: Dark background visible

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: add TailwindCSS with dark theme configuration"
```

---

## Task 4: Create Go Service Manager Interface

**Files:**
- Create: `internal/launchctl/types.go`
- Create: `internal/launchctl/manager.go`

**Step 1: Create types file**

Create: `internal/launchctl/types.go`

```go
package launchctl

// Service represents a LaunchAgent service
type Service struct {
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	Status      string            `json:"status"` // "running", "stopped", "unknown"
	PID         int               `json:"pid,omitempty"`
	Path        string            `json:"path"`
	Program     string            `json:"program,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
	RunAtLoad   bool              `json:"runAtLoad"`
	KeepAlive   bool              `json:"keepAlive"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	StdoutPath  string            `json:"stdoutPath,omitempty"`
	StderrPath  string            `json:"stderrPath,omitempty"`
	WorkingDir  string            `json:"workingDirectory,omitempty"`
}

// ScheduleConfig represents StartCalendarInterval
type ScheduleConfig struct {
	Minute  *int `json:"minute,omitempty"`
	Hour    *int `json:"hour,omitempty"`
	Day     *int `json:"day,omitempty"`
	Weekday *int `json:"weekday,omitempty"`
	Month   *int `json:"month,omitempty"`
}

// ServiceConfig is used for creating/updating services
type ServiceConfig struct {
	Label       string            `json:"label"`
	Program     string            `json:"program,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
	RunAtLoad   bool              `json:"runAtLoad"`
	KeepAlive   bool              `json:"keepAlive"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	StdoutPath  string            `json:"stdoutPath,omitempty"`
	StderrPath  string            `json:"stderrPath,omitempty"`
	WorkingDir  string            `json:"workingDirectory,omitempty"`
}
```

**Step 2: Create manager interface**

Create: `internal/launchctl/manager.go`

```go
package launchctl

// Manager defines the interface for managing LaunchAgents
type Manager interface {
	// List returns all services in the managed directory
	List() ([]Service, error)

	// Get returns a single service by name
	Get(name string) (*Service, error)

	// Start loads and starts a service
	Start(name string) error

	// Stop stops and unloads a service
	Stop(name string) error

	// Restart stops and starts a service
	Restart(name string) error

	// Create creates a new service with the given config
	Create(config *ServiceConfig) error

	// Update updates an existing service
	Update(name string, config *ServiceConfig) error

	// Delete removes a service
	Delete(name string) error

	// GetPlist returns the raw plist content
	GetPlist(name string) (string, error)

	// GetLogs returns stdout or stderr log content
	GetLogs(name string, logType string) (string, error)
}
```

**Step 3: Verify compilation**

Run:
```bash
go build ./...
```
Expected: No errors

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: define Service types and Manager interface"
```

---

## Task 5: Implement User LaunchAgent Manager

**Files:**
- Create: `internal/launchctl/user.go`
- Create: `internal/launchctl/user_test.go`

**Step 1: Write the test file**

Create: `internal/launchctl/user_test.go`

```go
package launchctl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserManager_List(t *testing.T) {
	m := NewUserManager()
	services, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	// Should return a slice (may be empty)
	if services == nil {
		t.Error("List() returned nil, expected empty slice")
	}
	t.Logf("Found %d services", len(services))
}

func TestUserManager_GetLaunchAgentsPath(t *testing.T) {
	m := &UserManager{}
	path := m.getLaunchAgentsPath()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents")

	if path != expected {
		t.Errorf("getLaunchAgentsPath() = %v, want %v", path, expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/launchctl/... -v
```
Expected: FAIL - NewUserManager not defined

**Step 3: Implement UserManager**

Create: `internal/launchctl/user.go`

```go
package launchctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

// UserManager manages user-level LaunchAgents
type UserManager struct {
	launchAgentsPath string
}

// NewUserManager creates a new UserManager
func NewUserManager() *UserManager {
	home, _ := os.UserHomeDir()
	return &UserManager{
		launchAgentsPath: filepath.Join(home, "Library", "LaunchAgents"),
	}
}

func (m *UserManager) getLaunchAgentsPath() string {
	return m.launchAgentsPath
}

// List returns all services in ~/Library/LaunchAgents
func (m *UserManager) List() ([]Service, error) {
	entries, err := os.ReadDir(m.launchAgentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Service{}, nil
		}
		return nil, fmt.Errorf("failed to read LaunchAgents directory: %w", err)
	}

	var services []Service
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.Get(name)
		if err != nil {
			continue // Skip files we can't parse
		}
		services = append(services, *service)
	}

	return services, nil
}

// Get returns a single service by name
func (m *UserManager) Get(name string) (*Service, error) {
	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")

	data, err := os.ReadFile(plistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plist: %w", err)
	}

	var plistData map[string]interface{}
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return nil, fmt.Errorf("failed to parse plist: %w", err)
	}

	service := &Service{
		Name:   name,
		Path:   plistPath,
		Status: "unknown",
	}

	// Parse plist fields
	if label, ok := plistData["Label"].(string); ok {
		service.Label = label
	} else {
		service.Label = name
	}

	if program, ok := plistData["Program"].(string); ok {
		service.Program = program
	}

	if args, ok := plistData["ProgramArguments"].([]interface{}); ok {
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				service.Arguments = append(service.Arguments, s)
			}
		}
		if service.Program == "" && len(service.Arguments) > 0 {
			service.Program = service.Arguments[0]
		}
	}

	if runAtLoad, ok := plistData["RunAtLoad"].(bool); ok {
		service.RunAtLoad = runAtLoad
	}

	if keepAlive, ok := plistData["KeepAlive"].(bool); ok {
		service.KeepAlive = keepAlive
	}

	if stdout, ok := plistData["StandardOutPath"].(string); ok {
		service.StdoutPath = stdout
	}

	if stderr, ok := plistData["StandardErrorPath"].(string); ok {
		service.StderrPath = stderr
	}

	if workDir, ok := plistData["WorkingDirectory"].(string); ok {
		service.WorkingDir = workDir
	}

	if envVars, ok := plistData["EnvironmentVariables"].(map[string]interface{}); ok {
		service.Environment = make(map[string]string)
		for k, v := range envVars {
			if s, ok := v.(string); ok {
				service.Environment[k] = s
			}
		}
	}

	// Get running status via launchctl
	service.Status, service.PID = m.getServiceStatus(service.Label)

	return service, nil
}

func (m *UserManager) getServiceStatus(label string) (string, int) {
	cmd := exec.Command("launchctl", "list", label)
	output, err := cmd.Output()
	if err != nil {
		return "stopped", 0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "stopped", 0
	}

	// Parse the output: PID, Status, Label
	fields := strings.Fields(lines[1])
	if len(fields) >= 1 && fields[0] != "-" {
		var pid int
		fmt.Sscanf(fields[0], "%d", &pid)
		if pid > 0 {
			return "running", pid
		}
	}

	return "stopped", 0
}

// Start loads and starts a service
func (m *UserManager) Start(name string) error {
	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")
	cmd := exec.Command("launchctl", "load", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s", string(output))
	}
	return nil
}

// Stop stops and unloads a service
func (m *UserManager) Stop(name string) error {
	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")
	cmd := exec.Command("launchctl", "unload", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %s", string(output))
	}
	return nil
}

// Restart stops and starts a service
func (m *UserManager) Restart(name string) error {
	if err := m.Stop(name); err != nil {
		// Ignore stop errors, service might not be running
	}
	return m.Start(name)
}

// GetPlist returns the raw plist content
func (m *UserManager) GetPlist(name string) (string, error) {
	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read plist: %w", err)
	}
	return string(data), nil
}

// GetLogs returns stdout or stderr log content
func (m *UserManager) GetLogs(name string, logType string) (string, error) {
	service, err := m.Get(name)
	if err != nil {
		return "", err
	}

	var logPath string
	switch logType {
	case "stdout":
		logPath = service.StdoutPath
	case "stderr":
		logPath = service.StderrPath
	default:
		return "", fmt.Errorf("invalid log type: %s", logType)
	}

	if logPath == "" {
		return "", fmt.Errorf("no %s log path configured", logType)
	}

	// Expand ~ to home directory
	if strings.HasPrefix(logPath, "~") {
		home, _ := os.UserHomeDir()
		logPath = filepath.Join(home, logPath[1:])
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("log file does not exist")
		}
		return "", fmt.Errorf("failed to read log: %w", err)
	}

	// Return last 1000 lines max
	lines := strings.Split(string(data), "\n")
	if len(lines) > 1000 {
		lines = lines[len(lines)-1000:]
	}

	return strings.Join(lines, "\n"), nil
}

// Create creates a new service
func (m *UserManager) Create(config *ServiceConfig) error {
	if config.Label == "" {
		return fmt.Errorf("label is required")
	}
	if config.Program == "" && len(config.Arguments) == 0 {
		return fmt.Errorf("program or arguments required")
	}

	plistPath := filepath.Join(m.launchAgentsPath, config.Label+".plist")
	if _, err := os.Stat(plistPath); err == nil {
		return fmt.Errorf("service already exists")
	}

	return m.writePlist(plistPath, config)
}

// Update updates an existing service
func (m *UserManager) Update(name string, config *ServiceConfig) error {
	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service does not exist")
	}

	// Stop the service first
	_ = m.Stop(name)

	return m.writePlist(plistPath, config)
}

// Delete removes a service
func (m *UserManager) Delete(name string) error {
	// Stop first
	_ = m.Stop(name)

	plistPath := filepath.Join(m.launchAgentsPath, name+".plist")
	if err := os.Remove(plistPath); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

func (m *UserManager) writePlist(path string, config *ServiceConfig) error {
	plistData := map[string]interface{}{
		"Label": config.Label,
	}

	if config.Program != "" {
		plistData["Program"] = config.Program
	}
	if len(config.Arguments) > 0 {
		plistData["ProgramArguments"] = config.Arguments
	}
	if config.RunAtLoad {
		plistData["RunAtLoad"] = true
	}
	if config.KeepAlive {
		plistData["KeepAlive"] = true
	}
	if config.StdoutPath != "" {
		plistData["StandardOutPath"] = config.StdoutPath
	}
	if config.StderrPath != "" {
		plistData["StandardErrorPath"] = config.StderrPath
	}
	if config.WorkingDir != "" {
		plistData["WorkingDirectory"] = config.WorkingDir
	}
	if len(config.Environment) > 0 {
		plistData["EnvironmentVariables"] = config.Environment
	}
	if config.Schedule != nil {
		schedule := map[string]interface{}{}
		if config.Schedule.Minute != nil {
			schedule["Minute"] = *config.Schedule.Minute
		}
		if config.Schedule.Hour != nil {
			schedule["Hour"] = *config.Schedule.Hour
		}
		if config.Schedule.Day != nil {
			schedule["Day"] = *config.Schedule.Day
		}
		if config.Schedule.Weekday != nil {
			schedule["Weekday"] = *config.Schedule.Weekday
		}
		if config.Schedule.Month != nil {
			schedule["Month"] = *config.Schedule.Month
		}
		if len(schedule) > 0 {
			plistData["StartCalendarInterval"] = schedule
		}
	}

	data, err := plist.MarshalIndent(plistData, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal plist: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
```

**Step 4: Add plist dependency**

Run:
```bash
go get howett.net/plist
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/launchctl/... -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: implement UserManager for LaunchAgents"
```

---

## Task 6: Create Wails App Bindings

**Files:**
- Modify: `app.go`
- Modify: `main.go`

**Step 1: Update app.go with service bindings**

Replace: `app.go`

```go
package main

import (
	"context"

	"launchpal/internal/launchctl"
)

// App struct
type App struct {
	ctx     context.Context
	manager *launchctl.UserManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		manager: launchctl.NewUserManager(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListServices returns all LaunchAgent services
func (a *App) ListServices() ([]launchctl.Service, error) {
	return a.manager.List()
}

// GetService returns a single service by name
func (a *App) GetService(name string) (*launchctl.Service, error) {
	return a.manager.Get(name)
}

// StartService starts a service
func (a *App) StartService(name string) error {
	return a.manager.Start(name)
}

// StopService stops a service
func (a *App) StopService(name string) error {
	return a.manager.Stop(name)
}

// RestartService restarts a service
func (a *App) RestartService(name string) error {
	return a.manager.Restart(name)
}

// GetPlist returns the raw plist content
func (a *App) GetPlist(name string) (string, error) {
	return a.manager.GetPlist(name)
}

// GetLogs returns log content
func (a *App) GetLogs(name string, logType string) (string, error) {
	return a.manager.GetLogs(name, logType)
}

// CreateService creates a new service
func (a *App) CreateService(config launchctl.ServiceConfig) error {
	return a.manager.Create(&config)
}

// UpdateService updates an existing service
func (a *App) UpdateService(name string, config launchctl.ServiceConfig) error {
	return a.manager.Update(name, &config)
}

// DeleteService deletes a service
func (a *App) DeleteService(name string) error {
	return a.manager.Delete(name)
}
```

**Step 2: Update main.go with macOS options**

Replace: `main.go`

```go
package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	app := NewApp()

	// Create application menu
	appMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
	}

	err := wails.Run(&options.App{
		Title:     "LaunchPal",
		Width:     1024,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Menu:             appMenu,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
			},
			Appearance: mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "LaunchPal",
				Message: "LaunchAgent Manager for macOS\n© 2025",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

**Step 3: Update go.mod module name**

Run:
```bash
go mod edit -module launchpal
go mod tidy
```

**Step 4: Test build**

Run:
```bash
wails build
```
Expected: Build succeeds

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add Wails app bindings for service management"
```

---

## Task 7: Create Main Layout Component

**Files:**
- Create: `frontend/layouts/default.vue`
- Create: `frontend/components/Sidebar.vue`
- Create: `frontend/components/StatusBar.vue`
- Modify: `frontend/app.vue`

**Step 1: Create Sidebar component**

Create: `frontend/components/Sidebar.vue`

```vue
<template>
  <aside class="w-16 bg-surface-500 flex flex-col items-center py-4 border-r border-surface-100">
    <nav class="flex flex-col gap-2">
      <NuxtLink
        to="/"
        class="p-3 rounded-lg transition-colors"
        :class="route.path === '/' ? 'bg-primary-600 text-white' : 'text-gray-400 hover:bg-surface-200 hover:text-white'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </NuxtLink>
    </nav>

    <div class="mt-auto">
      <NuxtLink
        to="/settings"
        class="p-3 rounded-lg transition-colors"
        :class="route.path === '/settings' ? 'bg-primary-600 text-white' : 'text-gray-400 hover:bg-surface-200 hover:text-white'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      </NuxtLink>
    </div>
  </aside>
</template>

<script setup lang="ts">
const route = useRoute()
</script>
```

**Step 2: Create StatusBar component**

Create: `frontend/components/StatusBar.vue`

```vue
<template>
  <footer class="h-8 bg-surface-500 border-t border-surface-100 flex items-center justify-between px-4 text-xs text-gray-400">
    <div class="flex items-center gap-4">
      <span>{{ serviceCount }} services</span>
      <span v-if="runningCount > 0" class="flex items-center gap-1">
        <span class="w-2 h-2 bg-green-500 rounded-full"></span>
        {{ runningCount }} running
      </span>
    </div>
    <div>
      v0.1.0
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps<{
  serviceCount: number
  runningCount: number
}>()
</script>
```

**Step 3: Create default layout**

Create: `frontend/layouts/default.vue`

```vue
<template>
  <div class="h-screen flex flex-col">
    <!-- Title bar drag region -->
    <div class="h-8 wails-drag bg-surface-500 flex items-center justify-center border-b border-surface-100">
      <span class="text-xs text-gray-400">LaunchPal</span>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <Sidebar />
      <main class="flex-1 flex flex-col overflow-hidden">
        <slot />
      </main>
    </div>

    <StatusBar :service-count="serviceCount" :running-count="runningCount" />
  </div>
</template>

<script setup lang="ts">
const serviceCount = ref(0)
const runningCount = ref(0)

// Will be updated by child components
provide('updateCounts', (total: number, running: number) => {
  serviceCount.value = total
  runningCount.value = running
})
</script>
```

**Step 4: Update app.vue**

Replace: `frontend/app.vue`

```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
```

**Step 5: Test dev mode**

Run:
```bash
wails dev
```
Expected: App shows with sidebar and status bar

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: create main layout with sidebar and status bar"
```

---

## Task 8: Create Services List Page

**Files:**
- Create: `frontend/pages/index.vue`
- Create: `frontend/components/ServiceRow.vue`

**Step 1: Create Wails runtime types**

Create: `frontend/types/wails.d.ts`

```typescript
export interface Service {
  name: string
  label: string
  status: 'running' | 'stopped' | 'unknown'
  pid?: number
  path: string
  program?: string
  arguments?: string[]
  runAtLoad: boolean
  keepAlive: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
}

export interface ScheduleConfig {
  minute?: number
  hour?: number
  day?: number
  weekday?: number
  month?: number
}

export interface ServiceConfig {
  label: string
  program?: string
  arguments?: string[]
  runAtLoad: boolean
  keepAlive: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          ListServices(): Promise<Service[]>
          GetService(name: string): Promise<Service>
          StartService(name: string): Promise<void>
          StopService(name: string): Promise<void>
          RestartService(name: string): Promise<void>
          GetPlist(name: string): Promise<string>
          GetLogs(name: string, logType: string): Promise<string>
          CreateService(config: ServiceConfig): Promise<void>
          UpdateService(name: string, config: ServiceConfig): Promise<void>
          DeleteService(name: string): Promise<void>
        }
      }
    }
  }
}

export {}
```

**Step 2: Create ServiceRow component**

Create: `frontend/components/ServiceRow.vue`

```vue
<template>
  <div
    class="flex items-center px-4 py-3 hover:bg-surface-200 cursor-pointer border-b border-surface-100"
    :class="{ 'bg-surface-200': selected }"
    @click="$emit('select', service.name)"
  >
    <input
      type="checkbox"
      class="mr-4 rounded bg-surface-300 border-surface-50"
      :checked="checked"
      @click.stop
      @change="$emit('check', service.name)"
    />

    <!-- Status indicator -->
    <div class="w-20 flex items-center">
      <span
        class="w-3 h-3 rounded-full"
        :class="service.status === 'running' ? 'bg-green-500' : 'bg-red-500'"
      ></span>
    </div>

    <!-- Name and subtitle -->
    <div class="flex-1 min-w-0">
      <div class="font-medium text-gray-100 truncate">{{ service.label }}</div>
      <div class="text-xs text-gray-400 truncate">
        {{ service.status === 'running' ? `Running · PID ${service.pid}` : 'Stopped' }}
      </div>
    </div>

    <!-- Schedule -->
    <div class="w-32 text-sm text-gray-400 truncate">
      {{ scheduleText }}
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2 ml-4">
      <button
        class="p-2 rounded hover:bg-surface-50 text-gray-400 hover:text-white"
        :title="service.status === 'running' ? 'Stop' : 'Start'"
        @click.stop="toggleService"
      >
        <svg v-if="service.status === 'running'" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      <button
        class="p-2 rounded hover:bg-surface-50 text-gray-400 hover:text-red-400"
        title="Delete"
        @click.stop="$emit('delete', service.name)"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

const props = defineProps<{
  service: Service
  selected: boolean
  checked: boolean
}>()

const emit = defineEmits<{
  select: [name: string]
  check: [name: string]
  delete: [name: string]
  refresh: []
}>()

const scheduleText = computed(() => {
  if (props.service.runAtLoad) return 'RunAtLoad'
  if (props.service.schedule) {
    const s = props.service.schedule
    if (s.hour !== undefined && s.minute !== undefined) {
      return `Daily ${String(s.hour).padStart(2, '0')}:${String(s.minute).padStart(2, '0')}`
    }
  }
  return '-'
})

async function toggleService() {
  try {
    if (props.service.status === 'running') {
      await window.go.main.App.StopService(props.service.name)
    } else {
      await window.go.main.App.StartService(props.service.name)
    }
    emit('refresh')
  } catch (err) {
    console.error('Failed to toggle service:', err)
  }
}
</script>
```

**Step 3: Create services list page**

Create: `frontend/pages/index.vue`

```vue
<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="px-6 py-4 border-b border-surface-100">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-semibold text-gray-100">Services</h1>
        <div class="flex items-center gap-2">
          <button
            class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg text-sm font-medium transition-colors"
            @click="showCreateModal = true"
          >
            + New Service
          </button>
        </div>
      </div>

      <!-- Search -->
      <div class="mt-4">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search services..."
          class="w-full px-4 py-2 bg-surface-300 border border-surface-100 rounded-lg text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
        />
      </div>
    </header>

    <!-- Table header -->
    <div class="flex items-center px-4 py-2 bg-surface-400 border-b border-surface-100 text-xs text-gray-400 uppercase tracking-wider">
      <div class="w-8"></div>
      <div class="w-20">Status</div>
      <div class="flex-1">Name</div>
      <div class="w-32">Schedule</div>
      <div class="w-24 text-right">Actions</div>
    </div>

    <!-- Services list -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="flex items-center justify-center h-32 text-gray-400">
        Loading...
      </div>
      <div v-else-if="filteredServices.length === 0" class="flex items-center justify-center h-32 text-gray-400">
        No services found
      </div>
      <template v-else>
        <ServiceRow
          v-for="service in filteredServices"
          :key="service.name"
          :service="service"
          :selected="selectedService === service.name"
          :checked="checkedServices.has(service.name)"
          @select="selectedService = $event"
          @check="toggleCheck"
          @delete="confirmDelete"
          @refresh="loadServices"
        />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

const services = ref<Service[]>([])
const loading = ref(true)
const searchQuery = ref('')
const selectedService = ref<string | null>(null)
const checkedServices = ref(new Set<string>())
const showCreateModal = ref(false)

const updateCounts = inject('updateCounts') as (total: number, running: number) => void

const filteredServices = computed(() => {
  if (!searchQuery.value) return services.value
  const query = searchQuery.value.toLowerCase()
  return services.value.filter(s =>
    s.name.toLowerCase().includes(query) ||
    s.label.toLowerCase().includes(query)
  )
})

async function loadServices() {
  try {
    services.value = await window.go.main.App.ListServices()
    const running = services.value.filter(s => s.status === 'running').length
    updateCounts(services.value.length, running)
  } catch (err) {
    console.error('Failed to load services:', err)
  } finally {
    loading.value = false
  }
}

function toggleCheck(name: string) {
  if (checkedServices.value.has(name)) {
    checkedServices.value.delete(name)
  } else {
    checkedServices.value.add(name)
  }
}

async function confirmDelete(name: string) {
  if (confirm(`Are you sure you want to delete "${name}"?`)) {
    try {
      await window.go.main.App.DeleteService(name)
      await loadServices()
    } catch (err) {
      console.error('Failed to delete service:', err)
    }
  }
}

onMounted(loadServices)
</script>
```

**Step 4: Test dev mode**

Run:
```bash
wails dev
```
Expected: Services list displays with your LaunchAgents

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: create services list page with search and actions"
```

---

## Task 9: Create Service Detail View

**Files:**
- Create: `frontend/pages/services/[name].vue`
- Create: `frontend/components/ServiceSummary.vue`
- Create: `frontend/components/ServiceLogs.vue`

**Step 1: Create ServiceSummary component**

Create: `frontend/components/ServiceSummary.vue`

```vue
<template>
  <div class="p-6 space-y-4">
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Program</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.program || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Working Directory</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.workingDirectory || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Arguments</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.arguments?.join(' ') || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">PID</label>
        <p class="text-gray-100 mt-1">{{ service.pid || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Run At Load</label>
        <p class="text-gray-100 mt-1">{{ service.runAtLoad ? 'Yes' : 'No' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Keep Alive</label>
        <p class="text-gray-100 mt-1">{{ service.keepAlive ? 'Yes' : 'No' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stdout Path</label>
        <p class="text-gray-100 mt-1 font-mono text-sm truncate">{{ service.stdoutPath || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stderr Path</label>
        <p class="text-gray-100 mt-1 font-mono text-sm truncate">{{ service.stderrPath || '-' }}</p>
      </div>
    </div>

    <div v-if="service.environment && Object.keys(service.environment).length > 0">
      <label class="text-xs text-gray-400 uppercase tracking-wider">Environment Variables</label>
      <div class="mt-2 bg-surface-300 rounded p-3 font-mono text-sm">
        <div v-for="(value, key) in service.environment" :key="key" class="text-gray-100">
          {{ key }}={{ value }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

defineProps<{
  service: Service
}>()
</script>
```

**Step 2: Create ServiceLogs component**

Create: `frontend/components/ServiceLogs.vue`

```vue
<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <div class="flex items-center gap-4 p-4 border-b border-surface-100">
      <select
        v-model="logType"
        class="px-3 py-1 bg-surface-300 border border-surface-100 rounded text-gray-100 text-sm"
      >
        <option value="stdout">stdout</option>
        <option value="stderr">stderr</option>
      </select>
      <button
        class="px-3 py-1 bg-surface-300 hover:bg-surface-200 rounded text-gray-100 text-sm"
        @click="loadLogs"
      >
        Refresh
      </button>
      <label class="flex items-center gap-2 text-sm text-gray-400">
        <input v-model="autoScroll" type="checkbox" class="rounded bg-surface-300" />
        Auto-scroll
      </label>
    </div>

    <div ref="logsContainer" class="flex-1 overflow-y-auto p-4 bg-surface-500 font-mono text-xs">
      <pre v-if="logs" class="text-gray-300 whitespace-pre-wrap">{{ logs }}</pre>
      <p v-else-if="loading" class="text-gray-400">Loading...</p>
      <p v-else class="text-gray-400">{{ error || 'No logs available' }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  serviceName: string
}>()

const logType = ref<'stdout' | 'stderr'>('stdout')
const logs = ref('')
const loading = ref(false)
const error = ref('')
const autoScroll = ref(true)
const logsContainer = ref<HTMLElement | null>(null)

async function loadLogs() {
  loading.value = true
  error.value = ''
  try {
    logs.value = await window.go.main.App.GetLogs(props.serviceName, logType.value)
    if (autoScroll.value && logsContainer.value) {
      nextTick(() => {
        logsContainer.value!.scrollTop = logsContainer.value!.scrollHeight
      })
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load logs'
    logs.value = ''
  } finally {
    loading.value = false
  }
}

watch(logType, loadLogs)
onMounted(loadLogs)
</script>
```

**Step 3: Create service detail page**

Create: `frontend/pages/services/[name].vue`

```vue
<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Breadcrumb header -->
    <header class="px-6 py-4 border-b border-surface-100">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 text-sm">
          <NuxtLink to="/" class="text-primary-400 hover:text-primary-300">Services</NuxtLink>
          <span class="text-gray-500">&gt;</span>
          <span class="text-gray-100">Service Details</span>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="p-2 rounded hover:bg-surface-200 text-gray-400 hover:text-white"
            :title="service?.status === 'running' ? 'Stop' : 'Start'"
            @click="toggleService"
          >
            <svg v-if="service?.status === 'running'" xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <button
            class="p-2 rounded hover:bg-surface-200 text-gray-400 hover:text-white"
            title="Restart"
            @click="restartService"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
          <button
            class="p-2 rounded hover:bg-surface-200 text-gray-400 hover:text-red-400"
            title="Delete"
            @click="deleteService"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Service name and status -->
      <div class="mt-4 flex items-center gap-3">
        <span
          class="w-4 h-4 rounded-full"
          :class="service?.status === 'running' ? 'bg-green-500' : 'bg-red-500'"
        ></span>
        <div>
          <h1 class="text-xl font-semibold text-gray-100">{{ service?.label }}</h1>
          <p class="text-sm text-gray-400">
            {{ service?.status === 'running' ? `Running · PID ${service?.pid}` : 'Stopped' }}
          </p>
        </div>
      </div>
    </header>

    <!-- Tabs -->
    <div class="flex border-b border-surface-100">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="px-6 py-3 text-sm font-medium transition-colors"
        :class="activeTab === tab.id
          ? 'text-primary-400 border-b-2 border-primary-400'
          : 'text-gray-400 hover:text-gray-100'"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab content -->
    <div class="flex-1 overflow-hidden">
      <ServiceSummary v-if="activeTab === 'summary' && service" :service="service" />
      <ServiceLogs v-else-if="activeTab === 'logs'" :service-name="name" />
      <div v-else-if="activeTab === 'inspect'" class="p-4 overflow-auto">
        <pre class="text-xs text-gray-300 font-mono whitespace-pre-wrap">{{ plistContent }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

const route = useRoute()
const router = useRouter()
const name = route.params.name as string

const service = ref<Service | null>(null)
const plistContent = ref('')
const activeTab = ref('summary')

const tabs = [
  { id: 'summary', label: 'Summary' },
  { id: 'logs', label: 'Logs' },
  { id: 'inspect', label: 'Inspect' },
]

async function loadService() {
  try {
    service.value = await window.go.main.App.GetService(name)
    plistContent.value = await window.go.main.App.GetPlist(name)
  } catch (err) {
    console.error('Failed to load service:', err)
  }
}

async function toggleService() {
  if (!service.value) return
  try {
    if (service.value.status === 'running') {
      await window.go.main.App.StopService(name)
    } else {
      await window.go.main.App.StartService(name)
    }
    await loadService()
  } catch (err) {
    console.error('Failed to toggle service:', err)
  }
}

async function restartService() {
  try {
    await window.go.main.App.RestartService(name)
    await loadService()
  } catch (err) {
    console.error('Failed to restart service:', err)
  }
}

async function deleteService() {
  if (confirm(`Are you sure you want to delete "${name}"?`)) {
    try {
      await window.go.main.App.DeleteService(name)
      router.push('/')
    } catch (err) {
      console.error('Failed to delete service:', err)
    }
  }
}

onMounted(loadService)
</script>
```

**Step 4: Update ServiceRow to navigate to detail**

Modify: `frontend/components/ServiceRow.vue`

Add to the click handler:
```vue
@click="navigateTo(`/services/${service.name}`)"
```

**Step 5: Test navigation**

Run:
```bash
wails dev
```
Expected: Clicking a service navigates to detail view

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: create service detail page with summary, logs, and inspect tabs"
```

---

## Task 10: Implement Backup System

**Files:**
- Create: `internal/backup/backup.go`
- Create: `internal/backup/backup_test.go`
- Modify: `app.go`

**Step 1: Write test file**

Create: `internal/backup/backup_test.go`

```go
package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupManager_GetBackupDir(t *testing.T) {
	m := NewBackupManager()
	dir := m.getBackupDir("com.test.service")

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".launchpal", "backups", "com.test.service")

	if dir != expected {
		t.Errorf("getBackupDir() = %v, want %v", dir, expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/backup/... -v
```
Expected: FAIL - NewBackupManager not defined

**Step 3: Implement BackupManager**

Create: `internal/backup/backup.go`

```go
package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxBackups = 10

// Backup represents a backup entry
type Backup struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// BackupManager handles backup operations
type BackupManager struct {
	baseDir string
}

// NewBackupManager creates a new BackupManager
func NewBackupManager() *BackupManager {
	home, _ := os.UserHomeDir()
	return &BackupManager{
		baseDir: filepath.Join(home, ".launchpal", "backups"),
	}
}

func (m *BackupManager) getBackupDir(serviceName string) string {
	return filepath.Join(m.baseDir, serviceName)
}

// Create creates a backup of the given plist file
func (m *BackupManager) Create(serviceName, plistPath string) (*Backup, error) {
	backupDir := m.getBackupDir(serviceName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now()
	backupID := timestamp.Format("20060102-150405")
	backupPath := filepath.Join(backupDir, backupID+".plist")

	// Copy the file
	src, err := os.Open(plistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Prune old backups
	m.pruneBackups(serviceName)

	return &Backup{
		ID:        backupID,
		Service:   serviceName,
		Timestamp: timestamp,
		Path:      backupPath,
	}, nil
}

// List returns all backups for a service
func (m *BackupManager) List(serviceName string) ([]Backup, error) {
	backupDir := m.getBackupDir(serviceName)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Backup{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".plist")
		timestamp, _ := time.Parse("20060102-150405", id)

		backups = append(backups, Backup{
			ID:        id,
			Service:   serviceName,
			Timestamp: timestamp,
			Path:      filepath.Join(backupDir, entry.Name()),
		})
	}

	// Sort by timestamp descending
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Get returns a specific backup
func (m *BackupManager) Get(serviceName, backupID string) (*Backup, error) {
	backupPath := filepath.Join(m.getBackupDir(serviceName), backupID+".plist")
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}

	return &Backup{
		ID:        backupID,
		Service:   serviceName,
		Timestamp: info.ModTime(),
		Path:      backupPath,
	}, nil
}

// GetContent returns the content of a backup
func (m *BackupManager) GetContent(serviceName, backupID string) (string, error) {
	backup, err := m.Get(serviceName, backupID)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(backup.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read backup: %w", err)
	}

	return string(data), nil
}

// Restore restores a backup to the original location
func (m *BackupManager) Restore(serviceName, backupID, targetPath string) error {
	backup, err := m.Get(serviceName, backupID)
	if err != nil {
		return err
	}

	// Create a backup of current state first
	if _, err := os.Stat(targetPath); err == nil {
		m.Create(serviceName, targetPath)
	}

	// Copy backup to target
	src, err := os.Open(backup.Path)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	return nil
}

func (m *BackupManager) pruneBackups(serviceName string) {
	backups, err := m.List(serviceName)
	if err != nil || len(backups) <= maxBackups {
		return
	}

	// Delete oldest backups
	for _, backup := range backups[maxBackups:] {
		os.Remove(backup.Path)
	}
}
```

**Step 4: Run tests**

Run:
```bash
go test ./internal/backup/... -v
```
Expected: PASS

**Step 5: Update app.go with backup methods**

Add to `app.go`:

```go
import "launchpal/internal/backup"

// Add to App struct
type App struct {
	ctx     context.Context
	manager *launchctl.UserManager
	backup  *backup.BackupManager
}

// Update NewApp
func NewApp() *App {
	return &App{
		manager: launchctl.NewUserManager(),
		backup:  backup.NewBackupManager(),
	}
}

// Add backup methods
func (a *App) ListBackups(serviceName string) ([]backup.Backup, error) {
	return a.backup.List(serviceName)
}

func (a *App) GetBackupContent(serviceName, backupID string) (string, error) {
	return a.backup.GetContent(serviceName, backupID)
}

func (a *App) RestoreBackup(serviceName, backupID string) error {
	service, err := a.manager.Get(serviceName)
	if err != nil {
		return err
	}
	return a.backup.Restore(serviceName, backupID, service.Path)
}
```

**Step 6: Update UpdateService to create backup**

Modify in `app.go`:

```go
func (a *App) UpdateService(name string, config launchctl.ServiceConfig) error {
	// Create backup before updating
	service, err := a.manager.Get(name)
	if err == nil {
		a.backup.Create(name, service.Path)
	}
	return a.manager.Update(name, &config)
}
```

**Step 7: Test build**

Run:
```bash
wails build
```
Expected: Build succeeds

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: implement backup system with auto-backup on update"
```

---

## Task 11: Build and Package Application

**Files:**
- Modify: `build/appicon.png`
- Modify: `wails.json`

**Step 1: Build release version**

Run:
```bash
wails build -clean
```
Expected: Creates `build/bin/launchpal.app`

**Step 2: Test the application**

Run:
```bash
open build/bin/launchpal.app
```
Expected: Application launches correctly

**Step 3: Commit final version**

```bash
git add -A
git commit -m "feat: complete LaunchPal v0.1.0"
```

---

## Summary

Tasks completed:
1. ✅ Initialize Wails project
2. ✅ Setup Nuxt 4 frontend
3. ✅ Install TailwindCSS with dark theme
4. ✅ Create Go Service Manager interface
5. ✅ Implement User LaunchAgent Manager
6. ✅ Create Wails App bindings
7. ✅ Create Main Layout component
8. ✅ Create Services List page
9. ✅ Create Service Detail view
10. ✅ Implement Backup system
11. ✅ Build and package application

**Future enhancements** (not in v0.1.0):
- Service creation modal
- Service edit form
- Menu bar tray icon
- System-level services support
- Settings page
