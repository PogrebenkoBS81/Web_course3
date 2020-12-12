package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

// IToken - token AAService interface.
type IToken interface {
	CreateToken(userId string) (*TokenDetails, error)
	ExtractTokenMetadata(*http.Request) (*AccessDetails, error)
}

// tokenService - IToken implementation.
type tokenService struct {
	logger logrus.FieldLogger
}

// NewToken - returns new IToken.
func NewToken(logger logrus.FieldLogger) IToken {
	return &tokenService{
		logger: logger.WithField("module", "tokenService"),
	}
}

// CreateToken - creates access and refresh jwt token.
func (t *tokenService) CreateToken(username string) (*TokenDetails, error) {
	entry := t.logger.WithField("func", "CreateToken")
	entry.Debugf("Creating token for:", username)

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 30).Unix() //expires after 30 min
	td.TokenUuid = uuid.New().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + username

	var err error
	//Creating Access Token
	entry.Debugf("Creating access token for:", username)
	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.TokenUuid
	atClaims["user_id"] = username
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodRS256, atClaims)
	block, rest := pem.Decode([]byte(os.Getenv("TOKEN_PRIVATE")))
	// If there is extra data after -----END PRIVATE KEY----- - throw an error.
	if len(rest) > 0 {
		return nil, fmt.Errorf("unexpected extra block length: %d", len(rest))
	}
	// Parse private key
	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, fmt.Errorf("at.ParsePKCS1PrivateKey(): %w", err)
	}
	// Sign access token with parsed private key.
	td.AccessToken, err = at.SignedString(pk)

	if err != nil {
		return nil, fmt.Errorf("at.SignedString(): %w", err)
	}

	//Creating Refresh Token
	entry.Debugf("Creating refresh token for:", username)
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + username

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = username
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodRS256, rtClaims)
	// Sign refresh token with parsed private key.
	td.RefreshToken, err = rt.SignedString(pk)

	if err != nil {
		return nil, fmt.Errorf("refresh at.SignedString(): %w", err)
	}

	entry.Debugf("Returning token details for:", username)
	return td, nil
}

// ExtractTokenMetadata - extracts token metadata from given request.
func (t *tokenService) ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	t.logger.WithField("func", "ExtractTokenMetadata").
		Debugf("Extracting token metadata for:", r.RemoteAddr)
	token, err := verifyToken(r)

	if err != nil {
		return nil, fmt.Errorf("verifyToken(): %w", err)
	}

	acc, err := extract(token)

	if err != nil {
		return nil, fmt.Errorf("extract(): %w", err)
	}

	return acc, nil
}
