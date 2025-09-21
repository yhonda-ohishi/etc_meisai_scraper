package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigration001Structure(t *testing.T) {
	migration := GetMigration001()

	assert.NotNil(t, migration)
	assert.Equal(t, "001_create_etc_tables", migration.ID())
	assert.NotEmpty(t, migration.Description())
}

func TestMigration001Implementation(t *testing.T) {
	migration := &Migration001CreateETCTables{}

	// Test interface implementation
	assert.Implements(t, (*Migration)(nil), migration)

	// Test ID and Description
	assert.Equal(t, "001_create_etc_tables", migration.ID())
	assert.Contains(t, migration.Description(), "ETC")
}