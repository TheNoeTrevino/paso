package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type OutputMode string

const (
	OutputModeHuman OutputMode = "human"
	OutputModeJSON  OutputMode = "json"
	OutputModeQuiet OutputMode = "quiet"
)

type OutputFormatter struct {
	mode   OutputMode
	writer io.Writer
}

func NewOutputFormatter(mode OutputMode) *OutputFormatter {
	return &OutputFormatter{
		mode:   mode,
		writer: os.Stdout,
	}
}

func (f *OutputFormatter) Print(message string) {
	if f.mode == OutputModeQuiet {
		return
	}
	fmt.Fprintln(f.writer, message)
}

func (f *OutputFormatter) PrintJSON(data interface{}) error {
	if f.mode == OutputModeQuiet {
		return nil
	}
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *OutputFormatter) PrintID(id string) {
	if f.mode == OutputModeHuman {
		fmt.Fprintf(f.writer, "ID: %s\n", id)
	} else {
		fmt.Fprintln(f.writer, id)
	}
}
