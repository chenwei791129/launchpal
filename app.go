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
