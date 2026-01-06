package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderDatabaseSelectLayer renders the database selection list as a modal layer
// This displays saved connections, "Local SQLite", and "Create New Connection..." options
func (m Model) renderDatabaseSelectLayer() *lipgloss.Layer {
	// Calculate dynamic height based on number of databases
	// databases + local + create + title + help + error padding
	numItems := len(m.DatabasePicker.SavedDatabases) + 2
	dynamicHeight := numItems + 6 // Add space for title, help, error, and padding

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  70,
			height: dynamicHeight,
		},
		contentRenderer: func(width, height int) string {
			return m.renderDatabaseSelectContent()
		},
		boxStyle: components.CreateInputBoxStyle,
	})
}

// renderDatabaseSelectContent generates the content for the database select picker
func (m Model) renderDatabaseSelectContent() string {
	var lines []string

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	lines = append(lines, titleStyle.Render("Select Database"))
	lines = append(lines, "")

	// Saved databases
	for i, db := range m.DatabasePicker.SavedDatabases {
		cursor := "  "
		if i == m.DatabasePicker.Cursor {
			cursor = "> "
		}
		// Render type chip
		typeChip := components.RenderDatabaseTypeChip(db.Type, m.Config.ColorScheme, "")
		displayStr := fmt.Sprintf("%s%s  %s", cursor, db.Name, typeChip)
		lines = append(lines, displayStr)
	}

	// Local SQLite option
	localIndex := len(m.DatabasePicker.SavedDatabases)
	cursor := "  "
	if localIndex == m.DatabasePicker.Cursor {
		cursor = "> "
	}
	typeChip := components.RenderDatabaseTypeChip("sqlite", m.Config.ColorScheme, "")
	lines = append(lines, fmt.Sprintf("%sLocal SQLite  %s", cursor, typeChip))

	// Create New Connection option
	createIndex := len(m.DatabasePicker.SavedDatabases) + 1
	cursor = "  "
	if createIndex == m.DatabasePicker.Cursor {
		cursor = "> "
	}
	lines = append(lines, fmt.Sprintf("%sCreate New Connection...", cursor))

	// Add error message if present
	if m.DatabasePicker.Err != nil {
		lines = append(lines, "")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ErrorFg))
		lines = append(lines, errorStyle.Render(fmt.Sprintf("Error: %s", m.DatabasePicker.Err.Error())))
	}

	// Add help text
	lines = append(lines, "")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	lines = append(lines, helpStyle.Render("↑/↓ navigate  enter select  d delete  esc cancel"))

	return strings.Join(lines, "\n")
}

// renderDatabaseCreateLayer renders the database connection creation form as a modal layer
func (m Model) renderDatabaseCreateLayer() *lipgloss.Layer {
	// Check if form is initialized
	if m.Forms.Form.DatabaseForm == nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ErrorFg))
		errorMsg := errorStyle.Render("Error: Form not initialized")
		return layers.CreateCenteredLayer(errorMsg, m.UIState.Width(), m.UIState.Height())
	}

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: dynamicPickerDimensions{
			itemCount: 4, // name, type, connection string, confirm
			hasFilter: false,
			minWidth:  60,
			maxWidth:  m.UIState.Width() * 3 / 4,
		},
		contentRenderer: func(width, height int) string {
			return m.renderDatabaseCreateContent()
		},
		boxStyle: components.CreateInputBoxStyle,
	})
}

// renderDatabaseCreateContent generates the content for the database create form
func (m Model) renderDatabaseCreateContent() string {
	formView := m.Forms.Form.DatabaseForm.View()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	title := titleStyle.Render("Create New Database Connection")

	// Show error if present
	var errorSection string
	if m.DatabasePicker.Err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ErrorFg))
		errorSection = errorStyle.Render(fmt.Sprintf("Error: %s", m.DatabasePicker.Err.Error()))
	}

	// Build content
	var contentParts []string
	contentParts = append(contentParts, title)
	contentParts = append(contentParts, "")
	contentParts = append(contentParts, formView)

	if errorSection != "" {
		contentParts = append(contentParts, "")
		contentParts = append(contentParts, errorSection)
	}

	return strings.Join(contentParts, "\n")
}

// renderDatabaseConnectConfirmLayer renders the database connection confirmation dialog as a modal layer
func (m Model) renderDatabaseConnectConfirmLayer() *lipgloss.Layer {
	if m.DatabasePicker.PendingConnection == nil {
		return nil
	}

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  50,
			height: 3, // message + blank line + options
		},
		contentRenderer: func(width, height int) string {
			cfg := m.DatabasePicker.PendingConnection
			return fmt.Sprintf("Connect to %s?\n\n[y]es  [n]o", cfg.Name)
		},
		boxStyle: components.DeleteConfirmBoxStyle,
	})
}

// renderDatabaseDeleteConfirmLayer renders the database deletion confirmation dialog
func (m Model) renderDatabaseDeleteConfirmLayer() *lipgloss.Layer {
	if m.DatabasePicker.PendingDeleteName == "" {
		return nil
	}

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  50,
			height: 3,
		},
		contentRenderer: func(width, height int) string {
			return fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", m.DatabasePicker.PendingDeleteName)
		},
		boxStyle: components.DeleteConfirmBoxStyle,
	})
}
