package huhforms

import (
	"strings"

	"charm.land/huh/v2"
)

// CreateDatabaseConnectionForm creates a huh form for adding a new database connection
func CreateDatabaseConnectionForm(
	name *string,
	connectionString *string,
	dbType *string,
	confirm *bool,
) *huh.Form {
	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Connection Name").
			Placeholder("e.g., Production DB, Dev Server").
			Value(name),

		huh.NewSelect[string]().
			Key("type").
			Title("Database Type").
			Options(
				huh.NewOption("PostgreSQL", "postgres"),
				huh.NewOption("SQLite", "sqlite"),
			).
			Value(dbType),

		huh.NewInput().
			Key("connection_string").
			Title("Connection String").
			Placeholder("e.g., postgres://user:pass@host:5432/dbname").
			Value(connectionString),

		huh.NewNote().
			TitleFunc(func() string {
				if connectionString != nil && strings.Contains(strings.ToLower(*connectionString), "password=") {
					return "⚠️  Connection strings with passwords are stored in plaintext.\nConsider using PostgreSQL .pgpass file instead."
				}
				return ""
			}, connectionString),

		huh.NewConfirm().
			Key("confirm").
			Title("Save this connection?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm),
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
