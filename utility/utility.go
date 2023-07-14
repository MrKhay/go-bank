package utility

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/mrkhay/gobank/storage"
	types "github.com/mrkhay/gobank/type"
)

type ApiError struct {
	Error string `json:"error"`
}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func permissionDenied(w http.ResponseWriter) {
	WriteJson(w, http.StatusBadGateway, ApiError{Error: "permission denied"})

}

func WithJWTAuth(handlerFunc http.HandlerFunc, s storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("x-jwt-token")
		token, err := ValidateJWT(tokenString)

		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return
		}

		userID, err := GetId(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)

		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if float64(account.AccountNumber) != claims["accountnumber"] {
			permissionDenied(w)
			return

		}

		handlerFunc(w, r)

	}
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	secreat := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secreat), nil
	})

}

func GetId(r *http.Request) (int, error) {
	idstr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idstr)

	if err != nil {
		return 0, fmt.Errorf("invalid id given %s", idstr)
	}

	return id, nil
}

func CreateJWT(account *types.Account) (string, error) {
	secreat := os.Getenv("JWT_SECRET")
	// create claims

	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountnumber": account.AccountNumber,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secreat))

}
