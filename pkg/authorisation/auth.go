package authorisation

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	UserAuth struct {
		Token string `json:"token"`
	}
)

var (
	jwtSecretKey = []byte("JWT_SECRET_KEY")
)

func VerifyToken(r *http.Request) (bool, error) {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		return false, fmt.Errorf("unauthorized") //	проверка заголовка
	}

	splitedToken := strings.Split(tokenString, " ")

	if len(splitedToken) < 2 || splitedToken[0] != "Bearer" { // проверка валидности формата
		return false, fmt.Errorf("unexpected token format\n\twant Authorization: Bearer <token>") // Authorization: Bearer <token>
	}

	token, err := jwt.Parse(tokenString[7:], func(t *jwt.Token) (interface{}, error) { //Parse проверяет подпись и возвращает проанализированный токен.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { // KeyFunc получит проанализированный токен и должен вернуть криптографический ключ для проверки подписи
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return jwtSecretKey, nil // если все ок, возвращаем ключ для проверки подписи
	})

	if err != nil {
		return false, err
	}

	if token.Valid { // проверка валидности токена, и его срока действия
		if expirationTime, _ := token.Claims.GetExpirationTime(); expirationTime.Unix() < time.Now().Unix() {
			return false, fmt.Errorf("token expired")
		}
		return true, nil
	}
	return false, nil

}

func GetToken(w http.ResponseWriter, r *http.Request) UserAuth {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{ //Claims для токена содержат время, когда действие истекает
		"exp": time.Now().Add(time.Minute * 30).Unix(),
	})

	tokenStr, _ := token.SignedString(jwtSecretKey) // подпись токена секретным ключом
	return UserAuth{Token: tokenStr}
}