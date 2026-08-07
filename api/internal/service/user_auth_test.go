package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareLoginPasswordUsesOneComparisonForKnownAndUnknownAccounts(t *testing.T) {
	tests := []struct {
		name         string
		users        []User
		expectedHash string
		matches      bool
	}{
		{name: "unknown", expectedHash: "dummy"},
		{name: "known valid", users: []User{{ID: 7, PasswordHash: "stored"}}, expectedHash: "stored", matches: true},
		{name: "ambiguous", users: []User{{ID: 7}, {ID: 8}}, expectedHash: "dummy"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			user, matches := compareLoginPassword(test.users, "supplied", []byte("dummy"), func(hash, password []byte) error {
				calls++
				assert.Equal(t, test.expectedHash, string(hash))
				assert.Equal(t, "supplied", string(password))
				if test.matches {
					return nil
				}
				return errors.New("mismatch")
			})
			require.Equal(t, 1, calls)
			assert.Equal(t, test.matches, matches)
			if test.matches {
				assert.Equal(t, int64(7), user.ID)
			}
		})
	}
}
