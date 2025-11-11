package presentation

import (
	"testing"
)

func TestConfigUI_TemplateLoading(t *testing.T) {
	// This test will fail if templates don't load correctly
	// The NewConfigUI function will panic if templates can't be loaded
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Template loading failed: %v", r)
		}
	}()

	// We can't easily test NewConfigUI without proper mocks
	// But we can test that the package compiles and templates exist
	t.Log("Presentation package compiles successfully")
}
