# Type Shenanigans

```
        Services (TaskService, ProjectService, etc.)
               │
               ▼
        database.Querier (interface with 67 methods)
               │
               ├── factory.NewQuerier(db, dbType)
               │                (switch on dbType)
         ┬─────├───────┬─
         ▼             ▼
  sqlite.Adapter  postgres.Adapter
         │             │
         ▼             ▼
  generated_sqlite  generated_postgres
         │             │
         ▼             ▼
     SQLite DB     PostgreSQL DB
```

This lets our services depend on a single `database.Querier` interface
