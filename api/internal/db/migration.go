package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"ariga.io/atlas-go-sdk/atlasexec"
)

//go:embed schemas/*.sql
var schemasFS embed.FS

func RunMigrations(dbURL string) error {
	// Extract the schemas subdirectory from the embedded FS
	schemasSubFS, err := fs.Sub(schemasFS, "schemas")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}

	// Create a working directory and write schema files to it
	workdir, err := atlasexec.NewWorkingDir()
	if err != nil {
		log.Fatalf("failed to load working directory: %v", err)
	}
	defer workdir.Close()

	// Write the schema.sql file from embedded FS to working directory
	schemaContent, err := fs.ReadFile(schemasSubFS, "schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	schemaPath := filepath.Join(workdir.Path(), "schema.sql")
	if err := os.WriteFile(schemaPath, schemaContent, 0644); err != nil {
		return fmt.Errorf("failed to write schema.sql: %w", err)
	}

	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		log.Fatalf("failed to initialize client: %v", err)
	}

	// Apply schema using schema apply command
	_, err = client.SchemaApply(context.Background(), &atlasexec.SchemaApplyParams{
		URL:         dbURL,
		To:          "file://schema.sql",
		DevURL:      "sqlite://file?mode=memory&_fk=1",
		AutoApprove: true,
	})
	if err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	fmt.Println("Schema applied successfully")
	return nil
}
