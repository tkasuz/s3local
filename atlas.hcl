// Atlas configuration for s3local SQLite database

// Define the SQLite environment
env "local" {
  src = "file://sqlc/schemas"
  dev = "sqlite://file?mode=memory"

  migration {
    dir = "file://internal/adapters/db/sqlite/migrations"
    format = golang-migrate
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = "file://sqlc/schemas"
  url = "sqlite://s3local.db"

  migration {
    dir = "file://internal/adapters/db/sqlite/migrations"
    format = golang-migrate
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
