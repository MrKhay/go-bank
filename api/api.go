package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrkhay/gobank/storage"
	"github.com/mrkhay/gobank/utility"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}
type ApiSuccess struct {
	Success string `json:"success"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			utility.WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}

type APISERVER struct {
	listenAddr string
	store      storage.Storage
}

func NewApiServer(listenAdr string, store storage.Storage) *APISERVER {
	return &APISERVER{
		listenAddr: listenAdr,
		store:      store,
	}
}

func (s *APISERVER) Run() {
	router := mux.NewRouter()

	// account
	router.HandleFunc("/test", makeHttpHandleFunc(s.testDB))
	router.HandleFunc("/topup", makeHttpHandleFunc(s.handleTopUp))
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/account/{id}", utility.WithJWTAuth(makeHttpHandleFunc(s.handleAccountWithID), s.store))

	// transactions
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))
	router.HandleFunc("/transactions", makeHttpHandleFunc(s.handleGetTransactions))
	router.HandleFunc("/transactions/{id}", makeHttpHandleFunc(s.handleGetUserTransactions))

	log.Println("JSON API SERVER running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)

}
