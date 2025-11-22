// Atlas configuration for s3local SQLite database

// Define the SQLite environment
env "local" {
  src = "file://schemas"
  dev = "sqlite://file?mode=memory&_fk=1"
}
