package token

import (
	"errors"
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

func GenerateShareToken(username string, audios map[string][]int, exp time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	claims["iat"] = time.Now().Unix()
	if exp.Milliseconds() > 0 {
		claims["exp"] = time.Now().Add(exp).Unix()
	}
	claims["username"] = username
	claims["audios"] = audios
	claims["type"] = "share"

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
	iat := int64(claims["iat"].(float64))
	if !storage.UserExists(username) {
		return "", errors.New("user not exist")
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

func ValidateShareToken(token string) (map[string][]int, error) {
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

	iat := int64(claims["iat"].(float64))

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

	audios, ok := claims["audios"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	return parseAudios(audios), nil
}

// return 0 for ok, 1 for no permission, 2 for authorization invalid
func CheckCoverPerms(token, catalog string) uint8 {
	_, err := ValidateUserToken(token)
	if err != nil {
		audios, err := ValidateShareToken(token)
		if err != nil {
			return 2
		} else {
			_, exists := audios[catalog]
			if exists {
				return 0
			} else {
				return 1
			}
		}
	} else {
		return 0
	}
}

// return 0 for ok, 1 for no permission, 2 for authorization invalid
func CheckAudioPerms(token, catalog string, track int) uint8 {
	_, err := ValidateUserToken(token)
	if err != nil {
		audios, err := ValidateShareToken(token)
		if err != nil {
			return 2
		} else {
			tracks, exists := audios[catalog]
			if !exists {
				return 1
			}
			if contains(tracks, track) {
				return 0
			}
			return 1
		}
	} else {
		return 0
	}
}

func contains(arr []int, el int) bool {
	for _, e := range arr {
		if e == el {
			return true
		}
	}
	return false
}

func parseAudios(e map[string]interface{}) map[string][]int {
	ret := make(map[string][]int)

	for k, v := range e {
		ar, ok := v.([]interface{})
		if !ok {
			return ret
		}
		arr := make([]int, 0)
		for _, va := range ar {
			track, ok := va.(float64)
			if !ok {
				return ret
			}
			arr = append(arr, int(track))
		}
		ret[k] = arr
	}

	return ret
}
