package auth

import (
	"activity_api/data_manager/cache"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

// IAuth - authorization AApi interface.
type IAuth interface {
	CreateAuth(string, *TokenDetails) error
	FetchAuth(string) (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
}

// NewAuth - returns new auth manager.
func NewAuth(client cache.ICacheManager, logger logrus.FieldLogger) IAuth {
	return &service{
		client: client,
		logger: logger.WithField("module", "service (IAuth)"),
	}
}

// service - service for auth.
type service struct {
	// Redis is used to store JWT tokens.
	client cache.ICacheManager
	logger logrus.FieldLogger
}

// AccessDetails - user access details.
type AccessDetails struct {
	TokenUuid string
	Username  string
}

// TokenDetails - JWT token details.
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

// CreateAuth - save token metadata to Redis
func (tk *service) CreateAuth(userId string, td *TokenDetails) error {
	entry := tk.logger.WithField("func", "CreateAuth")
	entry.Debug("Writing tokens to redis for admin:", userId)

	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	// Redis set expires with JWT access token.
	entry.Debugf("Setting access tokens uuid %s to redis for admin:", userId)
	err := tk.client.Set(td.TokenUuid, userId, at.Sub(now))

	if err != nil {
		return err
	}

	// Redis set expires with JWT refresh token.
	if err = tk.client.Set(td.RefreshUuid, userId, rt.Sub(now)); err != nil {
		return err
	}

	return nil
}

// FetchAuth - check the metadata saved
func (tk *service) FetchAuth(tokenUuid string) (string, error) {
	tk.logger.WithField("func", "FetchAuth").
		Debug("Fetching token from redis with uuid:", tokenUuid)

	userId, err := tk.client.Get(tokenUuid)

	if err != nil {
		return "", err
	}

	return userId, nil
}

// DeleteTokens - once a user row in the token table
func (tk *service) DeleteTokens(authD *AccessDetails) error {
	tk.logger.WithField("func", "DeleteTokens").
		Debug("Deleting tokens from redis:", authD)
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, authD.Username)
	//delete access token
	err := tk.client.Del(authD.TokenUuid)
	if err != nil {
		return err
	}
	//delete refresh token
	return tk.client.Del(refreshUuid)
}

// DeleteRefresh - deletes refresh token from redis.
func (tk *service) DeleteRefresh(refreshUuid string) error {
	tk.logger.WithField("func", "DeleteRefresh").
		Debug("Deleting refresh token from redis:", refreshUuid)

	return tk.client.Del(refreshUuid)
}
