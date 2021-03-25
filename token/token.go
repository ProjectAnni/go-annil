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

func ValidateShareToken(token string) (map[string][]uint8, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || alg != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	iat, ok := claims["iat"].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	date, err := storage.RegisterDate(username)
	if err != nil {
		return nil, err
	}

	// Token issued before user register
	if date.After(time.Unix(iat, 0)) {
		return nil, fmt.Errorf("invalid iat")
	}

	if claims["type"].(string) != "share" {
		return nil, fmt.Errorf("invalid type")
	}

	audios, ok := claims["audios"].(map[string][]uint8)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	return audios, nil
}

func CheckCoverPerms(token, catalog string) bool {
	_, err := ValidateUserToken(token)
	if err != nil {
		audios, err := ValidateShareToken(token)
		if err != nil {
			return false
		} else {
			_, exists := audios[catalog]
			return exists
		}
	} else {
		return true
	}
}

func CheckAudioPerms(token, catalog string, track uint8) bool {
	_, err := ValidateUserToken(token)
	if err != nil {
		audios, err := ValidateShareToken(token)
		if err != nil {
			return false
		} else {
			tracks, exists := audios[catalog]
			if !exists {
				return false
			}
			return contains(tracks, track)
		}
	} else {
		return true
	}
}

func contains(arr []uint8, el uint8) bool {
	for _, e := range arr {
		if e == el {
			return true
		}
	}
	return false
}
