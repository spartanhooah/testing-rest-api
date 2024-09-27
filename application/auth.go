package application

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spartanhooah/testing-rest-api/data"
	"net/http"
	"strings"
	"time"
)

var jwtTokenExpiry = time.Minute * 15
var RefreshTokenExpiry = time.Hour * 24

type TokenPairs struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (app *Application) generateTokenPair(user *data.User) (TokenPairs, error) {
	// create token
	token := jwt.New(jwt.SigningMethodHS256)

	// set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()

	if user.IsAdmin == 1 {
		claims["admin"] = true
	} else {
		claims["admin"] = false
	}

	// create the signed token
	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))

	if err != nil {
		return TokenPairs{}, err
	}

	// create the refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["exp"] = time.Now().Add(RefreshTokenExpiry).Unix()

	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))

	if err != nil {
		return TokenPairs{}, err
	}

	return TokenPairs{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}

func (app *Application) getTokenFromHeaderAndVerify(resp http.ResponseWriter, req *http.Request) (string, *Claims, error) {
	// add a header
	resp.Header().Add("Vary", "Authorization")

	// get the auth header
	authHeader := req.Header.Get("Authorization")

	// sanity check
	if authHeader == "" {
		return "", nil, errors.New("Missing Authorization header")
	}

	// split the header on spaces
	headerParts := strings.Split(authHeader, " ")

	if len(headerParts) != 2 {
		return "", nil, errors.New("Invalid Authorization header")
	}

	// check to see if the word "Bearer" is in the first part of the header
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("Unauthorized: no Bearer")
	}

	token := headerParts[1]

	// declare an empty Claims variable
	claims := &Claims{}

	// parse the token with our claims (we read into claims), using our secret (from the receiver)
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		// validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(app.JWTSecret), nil
	})

	// check for an error; note that this catches expired tokens as well
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("Token is expired")
		}

		return "", nil, err
	}

	// ensure we issued the token
	if claims.Issuer != app.Domain {
		return "", nil, errors.New("Incorrect issuer")
	}

	// token is valid
	return token, claims, nil
}
