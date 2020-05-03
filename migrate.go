package goadmin

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/pkg/errors"

	_ "github.com/golang-migrate/migrate/v4/source/aws_s3" /* nolint */
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/golang-migrate/migrate/v4/source/github_ee"
	_ "github.com/golang-migrate/migrate/v4/source/gitlab"
	_ "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/golang-migrate/migrate/v4/source/godoc_vfs"
	_ "github.com/golang-migrate/migrate/v4/source/google_cloud_storage"
	_ "github.com/golang-migrate/migrate/v4/source/stub"
)

func Migrate(config *DBConfig) error {
	driver, err := postgres.WithInstance(config.DB, &postgres.Config{
		MigrationsTable: MigrationsTable,
	})
	if err != nil {
		return errors.Wrap(err, "creating postgres instance failed")
	}

	mig, err := migrate.NewWithDatabaseInstance(
		config.MigrationsPath,
		config.Driver,
		driver,
	)

	if err != nil {
		return errors.Wrap(err, "creating database instance failed")
	}

	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "to migrate up failed")
	}

	return nil
}
