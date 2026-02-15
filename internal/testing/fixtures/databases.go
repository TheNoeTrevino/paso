package fixtures

import (
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
)

// DatabaseTestCase represents a parameterized test case for both SQLite and PostgreSQL
type DatabaseTestCase struct {
	Name    string
	DBType  database.DatabaseType
	Dialect Dialect
	SetupDB func(tb testing.TB) *sql.DB
}

// AllDatabases returns test cases for all supported databases.
// It includes SQLite and PostgreSQL (if available).
func AllDatabases() []DatabaseTestCase {
	return []DatabaseTestCase{
		{
			Name:    "SQLite",
			DBType:  database.SQLite,
			Dialect: SQLiteDialect(),
			SetupDB: SetupTestDB,
		},
		{
			// PostgreSQL is optional and will be skipped if not available.
			// SetupPostgresTestDB skips the test when PostgreSQL is not running.
			Name:    "PostgreSQL",
			DBType:  database.PostgreSQL,
			Dialect: PostgresDialect(),
			SetupDB: SetupPostgresTestDB,
		},
	}
}

// RunDatabaseTests is a helper that runs a test function for each database.
// The callback receives the database connection, dialect, and database.DatabaseType
// so services can be instantiated with the correct database type.
//
// Database cleanup is handled automatically by each SetupDB function via tb.Cleanup.
//
// Example:
//
//	func TestCreateLabel(t *testing.T) {
//	    fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
//	        svc, err := label.NewService(db, dbType, nil)
//	        require.NoError(t, err)
//	        // Your test code here
//	    })
//	}
func RunDatabaseTests(t *testing.T, testFunc func(t *testing.T, db *sql.DB, d Dialect, dbType database.DatabaseType)) {
	for _, tc := range AllDatabases() {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.SetupDB(t)
			if db == nil {
				// Test was skipped (e.g., PostgreSQL not available)
				return
			}

			testFunc(t, db, tc.Dialect, tc.DBType)
		})
	}
}

// RunDatabaseTestsWithSetup is like RunDatabaseTests but provides a setup phase for each database.
//
// Example:
//
//	func TestCreateLabel(t *testing.T) {
//	    fixtures.RunDatabaseTestsWithSetup(t,
//	        func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) any {
//	            projectID := fixtures.CreateTestProject(t, db, d, "Test Project")
//	            return projectID
//	        },
//	        func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType, setupData any) {
//	            // Test function using setupData
//	        },
//	    )
//	}
func RunDatabaseTestsWithSetup(
	t *testing.T,
	setupFunc func(t *testing.T, db *sql.DB, d Dialect, dbType database.DatabaseType) any,
	testFunc func(t *testing.T, db *sql.DB, d Dialect, dbType database.DatabaseType, setupData any),
) {
	for _, tc := range AllDatabases() {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.SetupDB(t)
			if db == nil {
				// Test was skipped
				return
			}

			setupData := setupFunc(t, db, tc.Dialect, tc.DBType)
			testFunc(t, db, tc.Dialect, tc.DBType, setupData)
		})
	}
}
