package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

func TestMigrateParams_Validate_EmptyDSN(t *testing.T) {
	p := &MigrateParams{}
	assert.ErrorIs(t, p.Validate(), errDSNRequired)
}

func TestMigrateParams_Validate_InvalidDirection(t *testing.T) {
	p := &MigrateParams{DSN: "postgres://localhost/test"}

	assert.ErrorIs(t, p.Validate(), errDirectionRequired)

	p.Direction = "sideways"
	assert.ErrorIs(t, p.Validate(), errDirectionRequired)
}

func TestMigrateParams_Validate_Up(t *testing.T) {
	p := &MigrateParams{DSN: "postgres://localhost/test", Direction: "up"}
	assert.NoError(t, p.Validate())
	assert.Equal(t, goadmin.DefaultMigrationsTable, p.MigrationsTable)
}

func TestMigrateParams_Validate_Down(t *testing.T) {
	p := &MigrateParams{DSN: "postgres://localhost/test", Direction: "down"}
	assert.NoError(t, p.Validate())
}

func TestMigrateParams_Validate_CustomTable(t *testing.T) {
	p := &MigrateParams{DSN: "postgres://localhost/test", Direction: "up", MigrationsTable: "custom"}
	assert.NoError(t, p.Validate())
	assert.Equal(t, "custom", p.MigrationsTable)
}

func TestCreateUserParams_Validate_EmptyDSN(t *testing.T) {
	p := &CreateUserParams{}
	assert.ErrorIs(t, p.Validate(), errDSNRequired)
}

func TestCreateUserParams_Validate_EmptyFields(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*CreateUserParams)
	}{
		{"empty login", func(p *CreateUserParams) { p.Login = "" }},
		{"empty name", func(p *CreateUserParams) { p.Name = "" }},
		{"empty role", func(p *CreateUserParams) { p.Role = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CreateUserParams{
				DSN:   "postgres://localhost/test",
				Login: "admin@example.com",
				Name:  "Admin",
				Role:  "owner",
			}
			tt.modify(p)
			assert.ErrorIs(t, p.Validate(), errFieldsRequired)
		})
	}
}

func TestCreateUserParams_Validate_Valid(t *testing.T) {
	p := &CreateUserParams{
		DSN:   "postgres://localhost/test",
		Login: "admin@example.com",
		Name:  "Admin",
		Role:  "owner",
	}
	assert.NoError(t, p.Validate())
	assert.Equal(t, goadmin.DefaultMigrationsTable, p.MigrationsTable)
}

func TestCreateUserParams_Validate_CustomTable(t *testing.T) {
	p := &CreateUserParams{
		DSN:             "postgres://localhost/test",
		Login:           "admin@example.com",
		Name:            "Admin",
		Role:            "owner",
		MigrationsTable: "custom_migrations",
	}
	assert.NoError(t, p.Validate())
	assert.Equal(t, "custom_migrations", p.MigrationsTable)
}

// --- resolvePassword ---

func TestResolvePassword_FlagValue(t *testing.T) {
	pw, err := resolvePassword("s3cr3t", nil)
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", pw)
}

func TestResolvePassword_EnvVar(t *testing.T) {
	t.Setenv("GOADMIN_PASSWORD", "from-env")

	pw, err := resolvePassword("", strings.NewReader(""))
	require.NoError(t, err)
	assert.Equal(t, "from-env", pw)
}

func TestResolvePassword_Stdin(t *testing.T) {
	pw, err := resolvePassword("", strings.NewReader("piped-password\n"))
	require.NoError(t, err)
	assert.Equal(t, "piped-password", pw)
}

func TestResolvePassword_StdinEmpty(t *testing.T) {
	_, err := resolvePassword("", strings.NewReader("\n"))
	assert.ErrorIs(t, err, errPasswordEmpty)
}

func TestResolvePassword_StdinEOF(t *testing.T) {
	_, err := resolvePassword("", strings.NewReader(""))
	assert.ErrorIs(t, err, errPasswordRequired)
}

func TestResolvePassword_FlagTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("GOADMIN_PASSWORD", "from-env")

	pw, err := resolvePassword("from-flag", strings.NewReader(""))
	require.NoError(t, err)
	assert.Equal(t, "from-flag", pw)
}
