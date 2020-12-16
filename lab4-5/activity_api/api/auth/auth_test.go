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
func checkFetch(t *testing.T, uuid, userID string) {
	id, err := auth.FetchAuth(uuid)

	if err != nil {
		t.Fatal(err)
	}

	if id != userID {
		t.Fatalf("userID %s doesn't match key %s userID %s", userID, uuid, id)
	}
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
	t.Run("CreateAuth_base", func(t *testing.T) {
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

	t.Run("FetchAuth_test", func(t *testing.T) {
		userID := uuid.New().String()
		token := getToken(userID)

		if err := auth.CreateAuth(userID, token); err != nil {
			t.Fatal(err)
		}

		checkFetch(t, token.TokenUuid, userID)
		checkFetch(t, token.RefreshUuid, userID)
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

	t.Run("DeleteTokens_test", func(t *testing.T) {
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
}
