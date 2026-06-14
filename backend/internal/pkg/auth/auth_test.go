package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("super-secret-123")
	require.NoError(t, err)
	assert.NotEqual(t, "super-secret-123", hash, "хеш не повинен дорівнювати паролю")

	assert.True(t, CheckPassword("super-secret-123", hash), "вірний пароль має пройти перевірку")
	assert.False(t, CheckPassword("wrong-password", hash), "невірний пароль не має пройти перевірку")
}

func TestGenerateAndParseToken(t *testing.T) {
	const secret = "test-secret"

	token, err := GenerateToken(42, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken(42, "correct-secret")
	require.NoError(t, err)

	_, err = ParseToken(token, "wrong-secret")
	assert.Error(t, err, "токен з невірним секретом має бути відхилений")
}
