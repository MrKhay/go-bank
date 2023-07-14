package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	t "github.com/mrkhay/gobank/type"
	util "github.com/mrkhay/gobank/utility"
)

func (s *APISERVER) handleAccount(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %v", r.Method)
}

func (s *APISERVER) handleAccountWithID(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {

		return s.handleGetAccountByID(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("invalid %v method", r.Method)

}

func (s *APISERVER) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {

	id, err := util.GetId(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)

	if err != nil {
		return err
	}

	return util.WriteJson(w, http.StatusOK, account)

}
func (s *APISERVER) testDB(w http.ResponseWriter, r *http.Request) error {

	res, err := s.store.TranscationTest()

	if err != nil {
		return err
	}

	return util.WriteJson(w, http.StatusOK, res)

}

func (s *APISERVER) handleTopUp(w http.ResponseWriter, r *http.Request) error {

	var req t.TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	err := s.store.TopUpAccount(&req)
	if err != nil {
		return err
	}

	return util.WriteJson(w, http.StatusOK, ApiSuccess{Success: fmt.Sprintf("account(%d) funded with $%s ", req.Account, req.Amount)})

}

func (s *APISERVER) handleGetAccount(w http.ResponseWriter, r *http.Request) error {

	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err

	}

	return util.WriteJson(w, http.StatusOK, accounts)

}

func (s *APISERVER) handleLogin(w http.ResponseWriter, r *http.Request) error {

	if r.Method != "POST" {
		return fmt.Errorf("invalid %v method", r.Method)
	}

	var req t.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByPasswordAndEmail(&req)

	if err != nil {
		return err
	}

	tokenString, err := util.CreateJWT(acc)

	if err != nil {
		return err
	}

	responce := struct {
		Account *t.Account `json:"account"`
		Token   *string    `json:"token"`
	}{
		Account: acc,
		Token:   &tokenString,
	}

	return util.WriteJson(w, http.StatusOK, responce)

}

// AccountResponse represents the response for the CreateAccount endpoint.
// swagger:response AccountResponse
type CreateAccountResonce struct {
	Account *t.Account `json:"account"`
	Token   *string    `json:"token"`
}

// CreateAccount returns account with token.
//
// swagger:route POST /account account CreateAccount
//
// Returns account with token..
//
// Responses:
//   200: AccountResponse
// 500:
//300

func (s *APISERVER) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {

	req := new(t.CreateAccountRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	account, err := t.NewAccount(req.FirstName, req.LastName, req.Email, req.Password)

	if err != nil {
		return err
	}

	if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.Password == "" {
		return fmt.Errorf("1 or more credentials are missing")
	}

	isInUse, err := s.store.CheckIfEmailExists(req.Email)

	if err != nil {
		return err
	}

	if isInUse {
		return fmt.Errorf("email address already in use")
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := util.CreateJWT(account)

	if err != nil {
		return err
	}

	responce := CreateAccountResonce{
		Account: account,
		Token:   &tokenString,
	}

	return util.WriteJson(w, http.StatusOK, responce)

}
func (s *APISERVER) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err := util.GetId(r)

	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return util.WriteJson(w, http.StatusOK, map[string]int{"deleted": id})
}
func (s *APISERVER) handleTransfer(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "POST" {
		var req t.TransferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		res, err := s.store.Transfer(&req)
		if err != nil {
			return err
		}

		return util.WriteJson(w, http.StatusOK, res)
	}

	return fmt.Errorf("invalid %v method", r.Method)

}
func (s *APISERVER) handleGetTransactions(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {

		t, err := s.store.GetTransactions()

		if err != nil {
			return err
		}
		return util.WriteJson(w, http.StatusOK, t)

	}

	return fmt.Errorf("invalid %v method", r.Method)

}

func (s *APISERVER) handleGetUserTransactions(w http.ResponseWriter, r *http.Request) error {

	id, err := util.GetId(r)
	if err != nil {
		return err
	}

	t, err := s.store.GetUserTransactions(id)

	if err != nil {
		return err
	}

	return util.WriteJson(w, http.StatusOK, &t)
}
