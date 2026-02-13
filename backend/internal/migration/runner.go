package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(databaseURL string) (uint, bool, error) {
	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return 0, false, err
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return 0, false, err
	}
	defer func() {
		_, _ = m.Close()
	}()

	changed := true
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return 0, false, err
		}
		changed = false
	}

	version, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, changed, err
	}

	return version, changed, nil
}

func CurrentVersion(databaseURL string) (uint, error) {
	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return 0, err
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return 0, err
	}
	defer func() {
		_, _ = m.Close()
	}()

	version, _, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

func resolveMigrationsPath() (string, error) {
	if configured := os.Getenv("MIGRATIONS_PATH"); configured != "" {
		if stat, err := os.Stat(configured); err == nil && stat.IsDir() {
			return filepath.Abs(configured)
		}
	}

	candidates := []string{"backend/migrations", "migrations"}
	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			return filepath.Abs(candidate)
		}
	}

	return "", errors.New("migrations path not found; set MIGRATIONS_PATH")
}
