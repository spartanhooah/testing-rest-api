package application

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spartanhooah/testing-rest-api/data"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

const refreshCookieName = "__Host-refresh_token"

type Credentials struct {
	Username string `json:"email"`
	Password string `json:"password"`
}

func (app *Application) authenticate(resp http.ResponseWriter, req *http.Request) {
	var creds Credentials

	// read a JSON payload
	err := app.readJSON(resp, req, &creds)

	if err != nil {
		app.errorJSON(resp, errors.New("Unauthorized"), http.StatusUnauthorized)
		return
	}

	// look up user by email address
	user, err := app.DB.GetUserByEmail(creds.Username)

	if err != nil {
		app.errorJSON(resp, errors.New("Unauthorized"), http.StatusUnauthorized)
		return
	}

	// check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))

	if err != nil {
		app.errorJSON(resp, errors.New("Unauthorized"), http.StatusUnauthorized)
		return
	}

	// generate token
	tokenPair, err := app.generateTokenPair(user)

	if err != nil {
		app.errorJSON(resp, errors.New("Unauthorized"), http.StatusUnauthorized)
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     refreshCookieName,
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(RefreshTokenExpiry),
		MaxAge:   int(RefreshTokenExpiry),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// send token to user
	_ = app.writeJSON(resp, http.StatusOK, tokenPair)
}

func (app *Application) refresh(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken := req.Form.Get("refresh_token")

	claims := &Claims{}

	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		return []byte(app.JWTSecret), nil
	})

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		app.errorJSON(resp, errors.New("refresh token does not to be renewed yet"), http.StatusTooEarly)
		return
	}

	userID, err := strconv.Atoi(claims.Subject)

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUser(userID)

	if err != nil {
		app.errorJSON(resp, errors.New("Unknown user"), http.StatusBadRequest)
		return
	}

	tokenPair, err := app.generateTokenPair(user)

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     refreshCookieName,
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(RefreshTokenExpiry),
		MaxAge:   int(RefreshTokenExpiry),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	_ = app.writeJSON(resp, http.StatusOK, tokenPair)
}

func (app *Application) refreshUsingCookie(resp http.ResponseWriter, req *http.Request) {
	for _, cookie := range req.Cookies() {
		if cookie.Name == refreshCookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
				return []byte(app.JWTSecret), nil
			})

			if err != nil {
				app.errorJSON(resp, err, http.StatusBadRequest)
				return
			}

			//if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
			//	app.errorJSON(resp, errors.New("refresh token does not to be renewed yet"), http.StatusTooEarly)
			//	return
			//}

			userID, err := strconv.Atoi(claims.Subject)

			if err != nil {
				app.errorJSON(resp, err, http.StatusBadRequest)
				return
			}

			user, err := app.DB.GetUser(userID)

			if err != nil {
				app.errorJSON(resp, errors.New("Unknown user"), http.StatusBadRequest)
				return
			}

			tokenPair, err := app.generateTokenPair(user)

			if err != nil {
				app.errorJSON(resp, err, http.StatusBadRequest)
				return
			}

			http.SetCookie(resp, &http.Cookie{
				Name:     refreshCookieName,
				Value:    tokenPair.RefreshToken,
				Path:     "/",
				Domain:   "localhost",
				Expires:  time.Now().Add(RefreshTokenExpiry),
				MaxAge:   int(RefreshTokenExpiry),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			// respond with JSON
			_ = app.writeJSON(resp, http.StatusOK, tokenPair)
			return
		}
	}

	app.errorJSON(resp, errors.New("Unauthorized"), http.StatusUnauthorized)
}

func (app *Application) allUsers(resp http.ResponseWriter, req *http.Request) {
	users, err := app.DB.AllUsers()

	if err != nil {
		app.errorJSON(resp, err, http.StatusInternalServerError)
		return
	}

	_ = app.writeJSON(resp, http.StatusOK, users)
}

func (app *Application) getUser(resp http.ResponseWriter, req *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(req, "userId"))

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUser(userId)

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	_ = app.writeJSON(resp, http.StatusOK, user)
}

func (app *Application) updateUser(resp http.ResponseWriter, req *http.Request) {
	var user data.User

	err := app.readJSON(resp, req, &user)

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	err = app.DB.UpdateUser(user)

	if err != nil {
		app.errorJSON(resp, err, http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

func (app *Application) deleteUser(resp http.ResponseWriter, req *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(req, "userId"))

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	err = app.DB.DeleteUser(userId)

	if err != nil {
		app.errorJSON(resp, err, http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

func (app *Application) createUser(resp http.ResponseWriter, req *http.Request) {
	var user data.User

	err := app.readJSON(resp, req, &user)

	if err != nil {
		app.errorJSON(resp, err, http.StatusBadRequest)
		return
	}

	_, err = app.DB.InsertUser(user)

	if err != nil {
		app.errorJSON(resp, err, http.StatusInternalServerError)
		return
	}

	_ = app.writeJSON(resp, http.StatusCreated, user)
}

func (app *Application) deleteRefreshCookie(resp http.ResponseWriter, req *http.Request) {
	http.SetCookie(resp, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	resp.WriteHeader(http.StatusAccepted)
}
