package token

import (
	"fmt"
	"github.com/SeraphJACK/go-annil/config"
	"github.com/SeraphJACK/go-annil/storage"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func GenerateUserToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	claims["iat"] = time.Now().Unix()
	claims["type"] = "user"
	claims["username"] = username
	claims["allowShare"] = storage.AllowShare(username)

	token.Claims = claims
	return token.SignedString([]byte(config.Cfg.Secret))
}

func GenerateShareToken(username string, audios map[string][]uint8, exp time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	claims["iat"] = time.Now().Unix()
	if exp.Milliseconds() > 0 {
		claims["exp"] = time.Now().Add(exp).Unix()
	}
	claims["username"] = username
	claims["audios"] = audios

	token.Claims = claims
	return token.SignedString([]byte(config.Cfg.Secret))
}

func ValidateUserToken(token string) (string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || alg != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Cfg.Secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}
	iat, ok := claims["iat"].(int64)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}
	date, err := storage.RegisterDate(username)
	if err != nil {
		return "", err
	}
	if claims["type"].(string) != "user" {
		return "", fmt.Errorf("invalid type")
	}
	// Token issued before user register
	if date.After(time.Unix(iat, 0)) {
		return "", fmt.Errorf("invalid iat")
	}
	return username, nil
}
