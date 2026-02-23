package standup

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// GenerateCmd returns the standup generate subcommand
func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a standup report",
		Long: `Generate a standup report showing logged entries within a time range.

Days and weeks are additive. For example, --days 3 --weeks 1 looks back 10 days.
Defaults to the last 7 days if no flags are specified.

Examples:
  # Generate report for the last week (default)
  paso standup generate

  # Last 3 days
  paso standup generate -d 3

  # Last 2 weeks
  paso standup generate -w 2

  # Last 1 week and 3 days (10 days total)
  paso standup generate -d 3 -w 1

  # JSON output
  paso standup generate -j`,
		RunE: runGenerate,
	}

	cmd.Flags().IntP("days", "d", 0, "Number of days to look back")
	cmd.Flags().IntP("weeks", "w", 0, "Number of weeks to look back (additive with --days)")
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (IDs only)")

	return cmd
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	days, _ := cmd.Flags().GetInt("days")
	weeks, _ := cmd.Flags().GetInt("weeks")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Default to 1 week if neither flag is provided
	if days == 0 && weeks == 0 {
		weeks = 1
	}

	// Use UTC so that query boundaries align with how both SQLite (CURRENT_TIMESTAMP stores
	// UTC) and PostgreSQL (lib/pq normalises to UTC over the wire) persist timestamps.
	now := time.Now().UTC()
	since := ComputeSince(now, days, weeks)
	until := now

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	projectID, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
			err.Error(),
			"Specify project ID with --project flag or associate branch with a project")
	}

	logs, err := cliInstance.App.StandupLogService.ListByProjectInRange(ctx, projectID, since, until)
	if err != nil {
		return formatter.Error(cli.ExitError, "GENERATE_ERROR", err.Error())
	}

	if quietMode {
		for _, id := range FormatGenerateQuiet(logs) {
			fmt.Println(id)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatGenerateJSON(logs, since, until))
	}

	if len(logs) == 0 {
		fmt.Printf("No standup logs found in the last ")
		if weeks > 0 {
			fmt.Printf("%d week(s) ", weeks)
		}
		if days > 0 {
			fmt.Printf("%d day(s) ", days)
		}
		fmt.Println()
		return nil
	}

	fmt.Print(FormatGenerateHuman(logs, since, until, cli.GetColorScheme()))
	return nil
}
