package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// OutputFormatter handles three output modes: JSON, quiet, and human-readable
type OutputFormatter struct {
	JSON  bool
	Quiet bool
}

// Success outputs successful operation result
func (f *OutputFormatter) Success(data interface{}) error {
	if f.Quiet {
		// Extract ID if possible
		if idGetter, ok := data.(interface{ GetID() int }); ok {
			fmt.Printf("%d\n", idGetter.GetID())
			return nil
		}
	}

	if f.JSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}

	// Human-readable format
	return f.prettyPrint(data)
}

// Error outputs error information
func (f *OutputFormatter) Error(code string, message string) error {
	return f.ErrorWithSuggestion(code, message, "")
}

// ErrorWithSuggestion outputs error information with an optional suggestion
func (f *OutputFormatter) ErrorWithSuggestion(code string, message string, suggestion string) error {
	if f.JSON {
		errData := map[string]interface{}{
			"code":    code,
			"message": message,
		}
		if suggestion != "" {
			errData["suggestion"] = suggestion
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": false,
			"error":   errData,
		})
	}

	// Human-readable error
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", message)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "💡 Suggestion: %s\n", suggestion)
	}
	return nil
}

// prettyPrint formats data for human-readable output
func (f *OutputFormatter) prettyPrint(data interface{}) error {
	// Default implementation - can be enhanced per data type
	fmt.Printf("%+v\n", data)
	return nil
}
