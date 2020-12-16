// https://github.com/victorsteven/gophercon-jwt-repo
package api

import (
	"activity_api/api/api_common"
	"activity_api/api/auth"
	"activity_api/common/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

// Login - login handler, checks request name and password,
// and if its valid, return access and refresh token to the user.
func (a *AApi) Login(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "Login")
	entry.Debug("Request from:", r.RemoteAddr)
	var req models.Admin

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Request from %s, admin name: %s", r.RemoteAddr, req.Username)
	// Check if admin with given name and password hash exists.
	if code, err := a.checkAdmin(&req); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			code,
			fmt.Sprintf("checkAdmin(): %v", err),
			a.logger,
		)

		return
	}
	// If all is fine - create auth token for admin.
	ts, err := a.token.CreateToken(req.Username)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateToken(): %v", err),
			a.logger,
		)

		return
	}

	if err := a.auth.CreateAuth(req.Username, ts); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateAuth(): %v", err),
			a.logger,
		)

		return
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	entry.Debugf("Responding to %s with tokens...", r.RemoteAddr)
	api_common.RespondWithJson(w, http.StatusOK, &tokens, a.logger)
}

// Logout - logouts user, deletes his tokens.
func (a *AApi) Logout(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "Logout")
	entry.Debugf("Request from %s", r.RemoteAddr)
	//If metadata is passed and the tokens valid, delete them from the redis store
	metadata, err := a.token.ExtractTokenMetadata(r)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("ExtractTokenMetadata(): %v", err),
			a.logger,
		)

		return
	}

	if metadata != nil {
		if err := a.auth.DeleteTokens(metadata); err != nil {
			entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
			api_common.RespondWithError(
				w,
				http.StatusBadRequest,
				fmt.Sprintf("DeleteTokens(): %v", err),
				a.logger,
			)

			return
		}
	}

	entry.Debugf("Responding to %s with OK...", r.RemoteAddr)
	api_common.RespondWithJson(w, http.StatusOK, nil, a.logger)
}

// Register - registers admin with data from JSON.
func (a *AApi) Register(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "Register")
	entry.Debug("Request from:", r.RemoteAddr)

	req := new(models.Admin)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}
	// Check if admin with given name exists in db.
	admin, err := a.sqlManager.GetAdmin(req.Username)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetAdmin(): %v", err),
			a.logger,
		)

		return
	}
	// If admin with given name already exists - return an error.
	if admin != nil {
		entry.Errorf("Respond to %s, admin with name %s already exists", r.RemoteAddr, req.Username)
		api_common.RespondWithError(
			w,
			http.StatusConflict,
			"admin with given name already exists",
			a.logger,
		)

		return
	}
	// Get hash of a password hash (salt is used).
	req.Hash, err = a.password.HashPassword(req.Hash)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("HashPassword(): %v", err),
			a.logger,
		)

		return
	}
	// Create admin if all is good.
	id, err := a.sqlManager.CreateAdmin(req)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateAdmin(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %d", r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusCreated, &models.ObjectID{ID: id}, a.logger)
}

// Unregister - deletes user from database.
func (a *AApi) Unregister(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "Unregister")
	entry.Debug("Request from:", r.RemoteAddr)

	metadata, err := a.token.ExtractTokenMetadata(r)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("ExtractTokenMetadata(): %v", err),
			a.logger,
		)

		return
	}

	if metadata == nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			"invalid user metadata",
			a.logger,
		)

		return
	}

	entry.Info("Deleting admin with name:", metadata.Username)
	_, err = a.sqlManager.DeleteAdmin(metadata.Username)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("DeleteAdmin(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Logout deleted user: %s (addr: %s)", metadata.Username, r.RemoteAddr)
	// If all is fine - logout deleted user.
	a.Logout(w, r)
}

// Refresh - refreshes user access token with refresh token.
func (a *AApi) Refresh(w http.ResponseWriter, r *http.Request) {
	mapToken := map[string]string{}
	decoder := json.NewDecoder(r.Body)

	entry := a.logger.WithField("func", "Refresh")
	entry.Debug("Request from:", r.RemoteAddr)

	if err := decoder.Decode(&mapToken); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}

	refreshToken := mapToken["refresh_token"]
	//verify the token
	token, err := auth.ParseToken(refreshToken)
	//if there is an error, the token must have expired
	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnauthorized,
			fmt.Sprintf("ParseToken(): %v", err),
			a.logger,
		)

		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		entry.Errorf("Respond error to %s, invalid token", r.RemoteAddr)
		api_common.RespondWithError(
			w,
			http.StatusUnauthorized,
			"invalid token",
			a.logger,
		)

		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		a.refresh(w, r, claims)
	} else {
		entry.Errorf("Respond error to %s, refresh expired", r.RemoteAddr)
		api_common.RespondWithError(
			w,
			http.StatusUnauthorized,
			"refresh expired",
			a.logger,
		)
	}
}

// refresh - Refresh helper.
func (a *AApi) refresh(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	entry := a.logger.WithField("func", "refresh")

	refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string

	if !ok {
		entry.Errorf("Respond error to %s, invalid refresh uuid", r.RemoteAddr)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			"invalid refresh uuid",
			a.logger,
		)

		return
	}

	userId, roleOk := claims["username"].(string)

	if roleOk == false {
		entry.Errorf("Respond error to %s, invalid username claims", r.RemoteAddr)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			"error getting username",
			a.logger,
		)

		return
	}

	entry.Debugf("Refreshing token for admin %s  (addr: %s)", userId, r.RemoteAddr)
	//Delete the previous Refresh Token
	err := a.auth.DeleteRefresh(refreshUuid)
	if err != nil { //if any goes wrong
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnauthorized,
			fmt.Sprintf("DeleteRefresh(): %v", err),
			a.logger,
		)

		return
	}
	//Create new pairs of refresh and access tokens
	entry.Debug("Creating token for admin:", userId)
	ts, err := a.token.CreateToken(userId)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusForbidden,
			fmt.Sprintf("CreateToken(): %v", err),
			a.logger,
		)

		return
	}

	//save the tokens metadata to redis
	entry.Debug("Creating auth for admin:", userId)
	saveErr := a.auth.CreateAuth(userId, ts)

	if saveErr != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, saveErr)
		api_common.RespondWithError(
			w,
			http.StatusForbidden,
			fmt.Sprintf("CreateAuth(): %v", err),
			a.logger,
		)

		return
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	entry.Debugf("Responding to %s (admin: %s) with refreshed tokens...", r.RemoteAddr, userId)
	api_common.RespondWithJson(w, http.StatusCreated, &tokens, a.logger)
}

func (a *AApi) checkAdmin(req *models.Admin) (int, error) {
	a.logger.WithField("func", "checkAdmin").Debug("Checking admin with name: ", req.Username)

	admin, err := a.sqlManager.GetAdmin(req.Username)

	if err != nil {
		return http.StatusUnprocessableEntity, err
	}

	if admin == nil {
		return http.StatusUnauthorized, errors.New("please provide valid login details")
	}

	if err = a.password.CheckPassword(req.Hash, admin.Hash); err != nil {
		return http.StatusUnauthorized, errors.New("please provide valid login details")
	}

	return -1, nil
}

// Decided not to use this check in Unregister
// SQL will return an error anyway if it doesn't exist

//admin, err := a.sqlManager.GetAdmin(metadata.Username)
//
//if err != nil {
//	api_common.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
//
//	return
//}
//
//if admin == nil {
//	api_common.RespondWithError(w, http.StatusNotFound, "admin doesn't exists")
//
//	return
//}
