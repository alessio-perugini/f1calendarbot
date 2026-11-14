package subscription

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestLoadMigrations(t *testing.T) {
	t.Parallel()

	migrations, err := loadMigrations()
	must.NoError(t, err)

	versions := make([]int, 0, len(migrations))
	for _, m := range migrations {
		must.NotEq(t, "", m.sql)
		versions = append(versions, m.version)
	}
	must.Eq(t, []int{1, 2}, versions)
	must.Eq(t, legacyVersion, versions[len(versions)-1])
}
