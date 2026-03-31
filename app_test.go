package main

import "testing"

func TestGetVersion_Default(t *testing.T) {
	app := NewApp()
	got := app.GetVersion()
	if got != "dev" {
		t.Errorf("GetVersion() = %q, want %q", got, "dev")
	}
}

func TestGetVersion_Injected(t *testing.T) {
	app := NewAppWithVersion("v1.6.0")
	got := app.GetVersion()
	if got != "v1.6.0" {
		t.Errorf("GetVersion() = %q, want %q", got, "v1.6.0")
	}
}
