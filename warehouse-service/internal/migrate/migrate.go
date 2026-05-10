// Package migrate применяет SQL-миграции golang-migrate.
package migrate

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Up применяет все миграции из migrationsDir.
func Up(dsn, migrationsDir string) error {
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return err
	}
	url := "file://" + filepath.ToSlash(abs)
	m, err := migrate.New(url, dsn)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
