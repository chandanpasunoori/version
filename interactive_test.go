package main

import (
	"strings"
	"testing"
)

func TestRunInteractiveSelection_ErrorOnEmptyChoices(t *testing.T) {
	_, err := runInteractiveSelection("Test", []string{})
	if err == nil {
		t.Error("Expected error for empty choices, but got none")
	}
	if err.Error() != "no choices available" {
		t.Errorf("Expected 'no choices available' error, got '%s'", err.Error())
	}
}

func TestListModel_Init(t *testing.T) {
	model := listModel{
		choices: []string{"option1", "option2"},
		title:   "Test Title",
	}
	
	cmd := model.Init()
	if cmd != nil {
		t.Error("Expected Init() to return nil")
	}
}

func TestListModel_View_EmptyChoices(t *testing.T) {
	model := listModel{
		choices: []string{},
		title:   "Test Title",
	}
	
	view := model.View()
	if !contains(view, "No items available") {
		t.Error("Expected view to contain 'No items available' for empty choices")
	}
}

func TestListModel_View_WithChoices(t *testing.T) {
	model := listModel{
		choices: []string{"option1", "option2"},
		title:   "Test Title",
		cursor:  0,
	}
	
	view := model.View()
	if !contains(view, "Test Title") {
		t.Error("Expected view to contain the title")
	}
	if !contains(view, "option1") {
		t.Error("Expected view to contain option1")
	}
	if !contains(view, "option2") {
		t.Error("Expected view to contain option2")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}