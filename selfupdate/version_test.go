//go:build unit

package selfupdate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rios0rios0/cliforge/selfupdate"
)

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	t.Run("should return -1 when v1 is older than v2", func(t *testing.T) {
		// given
		v1 := "1.0.0"
		v2 := "1.1.0"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, -1, result)
	})

	t.Run("should return 1 when v1 is newer than v2", func(t *testing.T) {
		// given
		v1 := "2.0.0"
		v2 := "1.9.9"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, 1, result)
	})

	t.Run("should return 0 when versions are equal", func(t *testing.T) {
		// given
		v1 := "1.2.3"
		v2 := "1.2.3"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, 0, result)
	})

	t.Run("should return -1 when v1 is dev", func(t *testing.T) {
		// given
		v1 := "dev"
		v2 := "0.0.1"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, -1, result)
	})

	t.Run("should return 1 when v2 is dev", func(t *testing.T) {
		// given
		v1 := "0.0.1"
		v2 := "dev"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, 1, result)
	})

	t.Run("should return 0 when versions have different lengths but are equal", func(t *testing.T) {
		// given
		v1 := "1.0"
		v2 := "1.0.0"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, 0, result)
	})

	t.Run("should return -1 when patch version is lower", func(t *testing.T) {
		// given
		v1 := "1.0.1"
		v2 := "1.0.2"

		// when
		result := selfupdate.CompareVersions(v1, v2)

		// then
		assert.Equal(t, -1, result)
	})
}
