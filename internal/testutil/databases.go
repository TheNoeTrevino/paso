package testutil

import (
	"database/sql"
	"testing"
)

// DatabaseType represents a supported database type
type DatabaseType string

const (
	SQLiteType     DatabaseType = "sqlite"
	PostgreSQLType DatabaseType = "postgres"
)

// DatabaseTestCase represents a parameterized test case for both SQLite and PostgreSQL
type DatabaseTestCase struct {
	Name     string
	DBType   DatabaseType
	Dialect  Dialect
	SetupDB  func(tb testing.TB) *sql.DB
	TearDown func(db *sql.DB)
}

// AllDatabases returns test cases for all supported databases
// It includes SQLite and PostgreSQL (if available)
func AllDatabases() []DatabaseTestCase {
	cases := []DatabaseTestCase{
		{
			Name:    "SQLite",
			DBType:  SQLiteType,
			Dialect: SQLiteDialect(),
			SetupDB: SetupTestDB,
			TearDown: func(db *sql.DB) {
				_ = db.Close()
			},
		},
	}

	// PostgreSQL is optional and will be skipped if not available
	// (The SetupPostgresTestDB function will skip the test if PostgreSQL is not running)
	cases = append(cases, DatabaseTestCase{
		Name:    "PostgreSQL",
		DBType:  PostgreSQLType,
		Dialect: PostgresDialect(),
		SetupDB: SetupPostgresTestDB,
		TearDown: func(db *sql.DB) {
			if db != nil {
				_ = db.Close()
			}
		},
	})

	return cases
}

// SqliteDatabasesOnly returns test cases for SQLite only
// Useful for tests that are only relevant to SQLite
func SqliteDatabasesOnly() []DatabaseTestCase {
	return []DatabaseTestCase{
		{
			Name:    "SQLite",
			DBType:  SQLiteType,
			Dialect: SQLiteDialect(),
			SetupDB: SetupTestDB,
			TearDown: func(db *sql.DB) {
				_ = db.Close()
			},
		},
	}
}

// PostgresqlDatabasesOnly returns test cases for PostgreSQL only
// Useful for tests that are only relevant to PostgreSQL
func PostgresqlDatabasesOnly() []DatabaseTestCase {
	return []DatabaseTestCase{
		{
			Name:    "PostgreSQL",
			DBType:  PostgreSQLType,
			Dialect: PostgresDialect(),
			SetupDB: SetupPostgresTestDB,
			TearDown: func(db *sql.DB) {
				if db != nil {
					_ = db.Close()
				}
			},
		},
	}
}

// RunDatabaseTests is a helper that runs a test function for each database
// Example:
//
//	func TestCreateLabel(t *testing.T) {
//	    testutil.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d testutil.Dialect) {
//	        // Your test code here
//	        label, err := svc.CreateLabel(ctx, req)
//	        require.NoError(t, err)
//	    })
//	}
func RunDatabaseTests(t *testing.T, testFunc func(t *testing.T, db *sql.DB, d Dialect)) {
	for _, tc := range AllDatabases() {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.SetupDB(t)
			if db == nil {
				// Test was skipped (e.g., PostgreSQL not available)
				return
			}
			defer tc.TearDown(db)

			testFunc(t, db, tc.Dialect)
		})
	}
}

// RunDatabaseTestsWithSetup is like RunDatabaseTests but provides setup/teardown for each database
// Example:
//
//	func TestCreateLabel(t *testing.T) {
//	    testutil.RunDatabaseTestsWithSetup(t,
//	        func(t *testing.T, db *sql.DB, d testutil.Dialect) any {
//	            // Setup for this database type
//	            projectID := testutil.CreateTestProject(t, db, d, "Test Project")
//	            return projectID
//	        },
//	        func(t *testing.T, db *sql.DB, d testutil.Dialect, setupData any) {
//	            // Test function using setupData
//	        },
//	    )
//	}
func RunDatabaseTestsWithSetup(
	t *testing.T,
	setupFunc func(t *testing.T, db *sql.DB, d Dialect) any,
	testFunc func(t *testing.T, db *sql.DB, d Dialect, setupData any),
) {
	for _, tc := range AllDatabases() {
		t.Run(tc.Name, func(t *testing.T) {
			db := tc.SetupDB(t)
			if db == nil {
				// Test was skipped
				return
			}
			defer tc.TearDown(db)

			setupData := setupFunc(t, db, tc.Dialect)
			testFunc(t, db, tc.Dialect, setupData)
		})
	}
}
