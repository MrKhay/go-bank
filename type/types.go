package types

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type TransferRequest struct {
	ToAccount   int       `json:"toAccount"`
	FromAccount int       `json:"fromAccount"`
	Amount      string    `json:"amount"`
	Date        time.Time `json:"date"`
}

type TranscationDetails struct {
	Sender   Account `json:"sender"`
	Receiver Account `json:"receiver"`
}

type TopUpRequest struct {
	Account int    `json:"acc_number"`
	Amount  string `json:"amount"`
}

type LoginRequest struct {
	Email   string `json:"email"`
	Pasword string `json:"password"`
}

// Account represents a account object.
// swagger:model
type Account struct {
	ID                int       `json:"-"`
	FirstName         string    `json:"firstname"`
	LastName          string    `json:"lastname"`
	AccountNumber     int64     `json:"acc_number"`
	Email             string    `json:"email"`
	EncryptedPassword string    `json:"-"`
	Balance           string    `json:"balance"`
	CreatedAt         time.Time `json:"createdAt"`
}

type Transcation struct {
	Id          uuid.UUID `json:"transaction_id"`
	Sen_acc     Account   `json:"sen_acc"`
	Rec_acc     Account   `json:"rec_acc"`
	Amount      string    `json:"amount"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Date        time.Time `json:"createdAt"`
}

func NewAccount(firstName, lastName, email, password string) (*Account, error) {
	encow, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	return &Account{
		FirstName:         firstName,
		LastName:          lastName,
		Email:             email,
		Balance:           "0",
		AccountNumber:     int64(rand.Intn(100000)),
		CreatedAt:         time.Now().UTC(),
		EncryptedPassword: string(encow),
	}, nil
}

func NewTransaction(s, r *int, amount, status, description string) (*Transcation, error) {

	uuid := uuid.New()
	sender := &Account{
		AccountNumber: int64(*s),
	}
	reciver := &Account{
		AccountNumber: int64(*r),
	}

	return &Transcation{
		Id:          uuid,
		Sen_acc:     *sender,
		Rec_acc:     *reciver,
		Status:      status,
		Amount:      amount,
		Description: description,
		Date:        time.Now().UTC(),
	}, nil
}
