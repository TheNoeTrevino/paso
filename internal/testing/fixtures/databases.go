package fixtures

import (
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
)

// DatabaseType represents a supported database type
type DatabaseType string

const (
	SQLiteType     DatabaseType = "sqlite"
	PostgreSQLType DatabaseType = "postgres"
)

// DBTypeToDatabase maps fixture DatabaseType to database.DatabaseType used by services.
func DBTypeToDatabase(dt DatabaseType) database.DatabaseType {
	switch dt {
	case PostgreSQLType:
		return database.PostgreSQL
	default:
		return database.SQLite
	}
}

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

// RunDatabaseTests is a helper that runs a test function for each database.
// The callback receives the database connection, dialect, and database.DatabaseType
// so services can be instantiated with the correct database type.
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
			defer tc.TearDown(db)

			testFunc(t, db, tc.Dialect, DBTypeToDatabase(tc.DBType))
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
			defer tc.TearDown(db)

			resolvedDBType := DBTypeToDatabase(tc.DBType)
			setupData := setupFunc(t, db, tc.Dialect, resolvedDBType)
			testFunc(t, db, tc.Dialect, resolvedDBType, setupData)
		})
	}
}
