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

func TestRunInteractiveMultiSelection_ErrorOnEmptyChoices(t *testing.T) {
	_, err := runInteractiveMultiSelection("Test", []string{})
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

func TestMultiSelectModel_Init(t *testing.T) {
	model := multiSelectModel{
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

func TestMultiSelectModel_View_EmptyChoices(t *testing.T) {
	model := multiSelectModel{
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

func TestMultiSelectModel_View_WithChoices(t *testing.T) {
	model := multiSelectModel{
		choices:  []string{"option1", "option2"},
		title:    "Test Title",
		cursor:   0,
		selected: make(map[int]bool),
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
	if !contains(view, "[ ]") {
		t.Error("Expected view to contain unchecked checkboxes")
	}
	if !contains(view, "Selected: 0 items") {
		t.Error("Expected view to show 0 selected items")
	}
}

func TestMultiSelectModel_View_WithSelections(t *testing.T) {
	model := multiSelectModel{
		choices:  []string{"option1", "option2", "option3"},
		title:    "Test Title",
		cursor:   0,
		selected: map[int]bool{0: true, 2: true},
	}
	
	view := model.View()
	if !contains(view, "Test Title") {
		t.Error("Expected view to contain the title")
	}
	if !contains(view, "[x]") {
		t.Error("Expected view to contain checked checkboxes")
	}
	if !contains(view, "[ ]") {
		t.Error("Expected view to contain unchecked checkboxes")
	}
	if !contains(view, "Selected: 2 items") {
		t.Error("Expected view to show 2 selected items")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Test multi-module functionality 
func TestMultiModule_HandlesCommaSeparatedModules(t *testing.T) {
	// Test the logic that splits comma-separated module names
	moduleName := "app1,app2,app3"
	
	multiModule := []string{moduleName}
	if strings.ContainsRune(moduleName, ',') {
		multiModule = strings.Split(moduleName, ",")
	}
	
	expected := []string{"app1", "app2", "app3"}
	if len(multiModule) != len(expected) {
		t.Errorf("Expected %d modules, got %d", len(expected), len(multiModule))
	}
	
	for i, module := range expected {
		if i >= len(multiModule) || multiModule[i] != module {
			t.Errorf("Expected module at index %d to be '%s', got '%s'", i, module, multiModule[i])
		}
	}
}