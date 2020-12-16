package auth

import (
	"activity_api/data_manager/cache"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"testing"
)

var (
	logger = &logrus.Logger{
		Level: logrus.FatalLevel,
	}

	cacheMock = cache.NewCacheManager(
		cache.ICacheMock,
		&cache.ICacheConfig{},
		context.Background(),
		logger,
	)

	auth = NewAuth(cacheMock, logger)
)

// getToken - token mock
func getToken(userID string) *TokenDetails {
	tokenUUID := uuid.New().String()

	return &TokenDetails{
		AtExpires:   0,
		RtExpires:   0,
		TokenUuid:   tokenUUID,
		RefreshUuid: tokenUUID + "++" + userID,
	}
}

// checkKey - checks if key exists in cache
func checkKey(cache cache.ICacheManager, key, value string) error {
	val, err := cache.Get(key)

	if err != nil {
		return err
	}

	if val != value {
		return fmt.Errorf("value %s doesn't match key %s value %s", value, key, val)
	}

	return nil
}

// checkFetch - checks FetchAuth
func checkFetch(uuid, userID string) error {
	id, err := auth.FetchAuth(uuid)

	if err != nil {
		return err
	}

	if id != userID {
		return fmt.Errorf("userID %s doesn't match key %s userID %s", userID, uuid, id)
	}

	return nil
}

// TestService_CreateAuth - tests CreateAuth
func TestService_CreateAuth(t *testing.T) {
	if err := cacheMock.Open(); err != nil {
		t.Fatal("I have no idea how this is happened")
	}
	defer func() {
		if err := cacheMock.Close(); err != nil {
			t.Fatal("I have no idea how this is happened")
		}
	}()
	// No need for subtests, but in 'future', if I will test another scenarios (something like border values),
	// they could be added as another t.Run("CreateAuth_blablabla"...
	t.Run("CreateAuth_test", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if err := auth.CreateAuth(userID, token); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.RefreshUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.TokenUuid, userID); err != nil {
			t.Fatal(err)
		}
	})
}

// TestService_FetchAuth - tests FetchAuth
func TestService_FetchAuth(t *testing.T) {
	if err := cacheMock.Open(); err != nil {
		t.Fatal("I have no idea how this is happened")
	}
	defer func() {
		if err := cacheMock.Close(); err != nil {
			t.Fatal("I have no idea how this is happened")
		}
	}()

	t.Run("FetchAuth_getExisted", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if err := auth.CreateAuth(userID, token); err != nil {
			t.Fatal(err)
		}

		if err := checkFetch(token.TokenUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := checkFetch(token.RefreshUuid, userID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FetchAuth_getInexistent", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if _, err := auth.FetchAuth(token.TokenUuid); err == nil {
			t.Fatal("no error for invalid input")
		}

		if _, err := auth.FetchAuth(token.RefreshUuid); err == nil {
			t.Fatal("no error for invalid input")
		}
	})
}

// TestService_DeleteTokens - tests DeleteTokens
func TestService_DeleteTokens(t *testing.T) {
	if err := cacheMock.Open(); err != nil {
		t.Fatal("I have no idea how this is happened")
	}
	defer func() {
		if err := cacheMock.Close(); err != nil {
			t.Fatal("I have no idea how this is happened")
		}
	}()

	t.Run("DeleteTokens_DeleteReal", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if err := auth.CreateAuth(userID, token); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.RefreshUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.TokenUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := auth.DeleteTokens(&AccessDetails{
			TokenUuid: token.TokenUuid,
			Username:  userID,
		}); err != nil {
			t.Fatal(err)
		}
		// If key exists after deletion - throw an error
		if err := checkKey(cacheMock, token.RefreshUuid, userID); err == nil {
			t.Fatal("token.RefreshUuid exists, but should not")
		}
		// If key exists after deletion - throw an error
		if err := checkKey(cacheMock, token.TokenUuid, userID); err == nil {
			t.Fatal("token.TokenUuid exists, but should not")
		}
	})

	t.Run("DeleteTokens_deleteInexistent", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)
		// No error should be returned for deleting inexistent key
		if err := auth.DeleteTokens(&AccessDetails{
			TokenUuid: token.TokenUuid,
			Username:  userID,
		}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestService_DeleteTokens - tests DeleteRefresh
func TestService_DeleteRefresh(t *testing.T) {
	if err := cacheMock.Open(); err != nil {
		t.Fatal("I have no idea how this is happened")
	}
	defer func() {
		if err := cacheMock.Close(); err != nil {
			t.Fatal("I have no idea how this is happened")
		}
	}()

	t.Run("DeleteRefresh_test", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if err := auth.CreateAuth(userID, token); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.RefreshUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := checkKey(cacheMock, token.TokenUuid, userID); err != nil {
			t.Fatal(err)
		}

		if err := auth.DeleteRefresh(token.RefreshUuid); err != nil {
			t.Fatal(err)
		}
		// If key exists after deletion - throw an error
		if err := checkKey(cacheMock, token.RefreshUuid, userID); err == nil {
			t.Fatal("token.RefreshUuid exists, but should not")
		}
		// This key shouldn't be touched
		if err := checkKey(cacheMock, token.TokenUuid, userID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteRefresh_deleteInexistent", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)
		// No error should be returned for deleting inexistent key
		if err := auth.DeleteRefresh(token.RefreshUuid); err != nil {
			t.Fatal(err)
		}
	})
}
