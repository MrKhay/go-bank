package storage

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
	t "github.com/mrkhay/gobank/type"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateAccount(*t.Account) error
	DeleteAccount(int) error
	UpdateAccount(*t.Account) error
	GetAccounts() ([]*t.Account, error)
	AccountQuerey
	Transaction
}

type AccountQuerey interface {
	GetAccountByID(int) (*t.Account, error)
	GetAccountByNumber(int) (*int, error)
	GetAccountByPasswordAndEmail(req *t.LoginRequest) (*t.Account, error)
	CheckIfEmailExists(email string) (bool, error)
}

type Transaction interface {
	TranscationTest() (bool, error)
	Transfer(req *t.TransferRequest) (*t.Transcation, error)
	TopUpAccount(req *t.TopUpRequest) error
	GetUserTransactions(acc_num int) ([]*t.Transcation, error)
	GetTransactions() ([]*t.Transcation, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable host=localhost port=5432"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db: db,
	}, nil

}

func (s *PostgresStorage) Init() error {

	return s.CreateAccountTable()

}
func (s *PostgresStorage) CreateAccountTable() error {

	querey := `CREATE TABLE IF NOT EXISTS accounts (
	id serial,
	first_name varchar(50),
	last_name varchar(50),
	acc_number serial unique,
	balance money,
	email varchar(50) ,
    password varchar(200),
	created_at timestamp
  )`

	_, err := s.db.Exec(querey)

	if err != nil {
		return err
	}

	querey = `CREATE TABLE IF NOT EXISTS transactions (
		transaction_id uuid primary key,
		sen_acc serial references accounts(acc_number),
		rec_acc serial references accounts(acc_number),
		amount money,
		description varchar(80),
		status varchar(20),
		date timestamp
		)`

	_, err = s.db.Exec(querey)

	if err != nil {
		return err
	}

	querey = `CREATE OR REPLACE VIEW transacationview AS
	SELECT t.transaction_id, t.amount, t.description,t.status,t.date,s.acc_number AS sender_acc,
	s.first_name AS sender_fn,s.last_name AS sender_ln, s.balance AS sender_balance,s.email AS
	sender_email,r.acc_number AS receiver_acc, r.first_name AS receiver_fn,r.last_name AS receiver_ln,
	r.balance AS receiver_balance,r.email AS receiver_email FROM transactions t JOIN accounts s ON t.sen_acc=s.acc_number
	 JOIN accounts r ON t.rec_acc=r.acc_number`

	_, err = s.db.Exec(querey)

	return err
}

func (s *PostgresStorage) GetAccountByPasswordAndEmail(req *t.LoginRequest) (*t.Account, error) {

	rows, err := s.db.Query("select * from accounts where email = $1", req.Email)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		acc, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		// validating password

		if err := bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(req.Pasword)); err != nil {

			return nil, fmt.Errorf("invalid password")
		} else {
			return acc, nil
		}

	}

	return nil, fmt.Errorf("accounts with email [ %s ] not found", req.Email)
}

func (s *PostgresStorage) CreateAccount(acc *t.Account) error {

	// begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err

	}

	query :=
		`insert into accounts
	(first_name, last_name, acc_number, balance, email, password, created_at)
	values($1,$2,$3,$4,$5,$6,$7)
	RETURNING id`

	res, err := tx.Exec(
		query,
		acc.FirstName,
		acc.LastName,
		acc.AccountNumber,
		acc.Balance,
		acc.Email,
		acc.EncryptedPassword,
		acc.CreatedAt)

	if err != nil {
		tx.Rollback()
		return err
	}

	i, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if i < 1 {
		return fmt.Errorf("account not found")
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
func (s *PostgresStorage) GetAccounts() ([]*t.Account, error) {

	rows, err := s.db.Query("select * from accounts")

	if err != nil {
		return nil, err
	}

	accounts := []*t.Account{}
	for rows.Next() {

		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *PostgresStorage) GetAccountByNumber(number int) (*int, error) {

	rows, err := s.db.Query(`select acc_number from accounts where acc_number = $1`, number)

	if err != nil {
		return nil, err
	}

	for rows.Next() {

		acc_num := 0
		err := rows.Scan(
			&acc_num)

		if err != nil {
			return nil, err
		} else {
			return &acc_num, nil
		}
	}

	return nil, fmt.Errorf("account with acc_number [ %d ] not found", number)
}

func (s *PostgresStorage) UpdateAccount(*t.Account) error {
	return nil
}
func (s *PostgresStorage) DeleteAccount(id int) error {

	_, err := s.db.Exec("delete from accounts where id = $1", id)

	if err != nil {
		return fmt.Errorf("account with id:{ %d } not found", id)
	}

	return nil
}
func (s *PostgresStorage) GetAccountByID(id int) (*t.Account, error) {

	rows, err := s.db.Query("select * from accounts where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	defer rows.Close()
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStorage) CheckIfEmailExists(email string) (bool, error) {

	rows, err := s.db.Query("select email from accounts where email = $1", email)

	if err != nil {
		return false, err
	}
	account := new(t.Account)
	for rows.Next() {

		err := rows.Scan(
			&account.Email,
		)

		if err != nil {
			return false, nil
		} else {
			return true, nil
		}
	}

	defer rows.Close()

	return false, nil
}

// transactions

func (s *PostgresStorage) TranscationTest() (bool, error) {

	// begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return false, err

	}

	name := "DFGHJK"
	// function
	res, err := tx.Exec(`delete from accounts where first_name = $1`, name)

	if err != nil {
		tx.Rollback()
		return false, err
	}

	r, _ := res.RowsAffected()

	if r < 1 {
		fmt.Println(r)
		return false, fmt.Errorf("account not found")
	}

	// commit the transaction
	err = tx.Commit()

	if err != nil {
		return false, err
	}

	return true, nil

}

func (s *PostgresStorage) GetTransactiobById(id *string) (*t.Transcation, error) {

	rows, err := s.db.Query(`select * from transacationview where transaction_id = $1`, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {

		t, err := scanIntoTransaction(rows)

		if err != nil {
			return nil, err
		} else {
			return t, nil
		}
	}

	return nil, fmt.Errorf("transaction with id [ %d ] not found", id)
}
func (s *PostgresStorage) Transfer(req *t.TransferRequest) (*t.Transcation, error) {

	// begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		fmt.Println("1")
		return nil, err

	}

	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		fmt.Println("2")
		return nil, err
	}

	// remove from sender account
	res, err := tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE acc_number = $2 AND balance > $1`, amount, req.FromAccount)

	if err != nil {
		fmt.Println("3")
		tx.Rollback()
		return nil, err
	}

	r, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if r < 1 {
		fmt.Println("4")
		tx.Rollback()
		return nil, fmt.Errorf("insufficient fund or invalid accound number")
	}

	// add to receiver account
	res, err = tx.Exec(`UPDATE accounts SET balance = $1 + balance WHERE acc_number = $2`, amount, req.ToAccount)

	if err != nil {
		fmt.Println("6")
		tx.Rollback()
		return nil, err
	}

	r, err = res.RowsAffected()

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	fmt.Println(r)

	if r < 1 {
		fmt.Println(r)
		tx.Rollback()
		return nil, fmt.Errorf("something went wrong")
	}

	transaction, err := t.NewTransaction(&req.FromAccount, &req.ToAccount, req.Amount, "Credit", "Bank Transfer")
	if err != nil {
		fmt.Println("9")
		tx.Rollback()
		return nil, err
	}

	id, err := s.AddTransaction(transaction)

	if err != nil {
		fmt.Println("10")
		tx.Rollback()
		return nil, err
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Println("12")
		tx.Rollback()
		return nil, err
	}

	t, err := s.GetTransactiobById(id)

	if err != nil {
		fmt.Println("11")
		tx.Rollback()
		return nil, err
	}

	return t, nil

}

func (s *PostgresStorage) AddTransaction(t *t.Transcation) (*string, error) {

	query := `
	INSERT INTO transactions
	(transaction_id,sen_acc,rec_acc,amount,description,status,date)
	VALUES($1,$2,$3,$4,$5,$6,$7)
    RETURNING transaction_id`

	row, err := s.db.Query(
		query,
		t.Id,
		t.Sen_acc.AccountNumber,
		t.Rec_acc.AccountNumber,
		t.Amount,
		t.Description,
		t.Status,
		t.Date)

	if err != nil {
		return nil, err
	}

	var id string
	for row.Next() {

		err := row.Scan(
			&id,
		)

		if err != nil {
			return nil, err
		} else {
			return &id, nil

		}
	}

	return nil, err
}
func (s *PostgresStorage) TopUpAccount(req *t.TopUpRequest) error {

	// begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// function
	res, err := tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE acc_number = $2`, req.Amount, req.Account)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf(tx.Rollback().Error())
	}

	r, _ := res.RowsAffected()

	if r < 1 {
		fmt.Println(r)
		return fmt.Errorf("account not found")
	}

	// commit the transaction
	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil

}

func (s *PostgresStorage) GetUserTransactions(acc_num int) ([]*t.Transcation, error) {

	// function
	rows, err := s.db.Query(`SELECT * FROM transacationview WHERE sender_acc = $1 OR receiver_acc = $1 `, acc_num)

	if err != nil {
		return nil, err
	}

	transactions := []*t.Transcation{}
	for rows.Next() {

		transcation, err := scanIntoTransaction(rows)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transcation)
	}
	return transactions, nil
}

func (s *PostgresStorage) GetTransactions() ([]*t.Transcation, error) {

	rows, err := s.db.Query("select * from transacationview")

	if err != nil {
		return nil, err
	}

	transactions := []*t.Transcation{}
	for rows.Next() {

		transcation, err := scanIntoTransaction(rows)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transcation)
	}
	return transactions, nil
}
func DropTable(db *sql.DB, n string) error {

	if _, err := db.Query(`DROP TABLE $1`, n); err != nil {
		return err
	}
	return nil
}

func scanIntoAccount(rows *sql.Rows) (*t.Account, error) {

	account := new(t.Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.AccountNumber,
		&account.Balance,
		&account.Email,
		&account.EncryptedPassword,
		&account.CreatedAt,
	)

	return account, err

}

func scanIntoTransaction(rows *sql.Rows) (*t.Transcation, error) {
	s := new(t.Account)
	r := new(t.Account)
	tran := new(t.Transcation)

	// s.Balance = sb
	// r.Balance = rb

	err := rows.Scan(
		&tran.Id,
		&tran.Amount,
		&tran.Description,
		&tran.Status,
		&tran.Date,
		&s.AccountNumber,
		&s.FirstName,
		&s.LastName,
		&s.Balance,
		&s.Email,
		&r.AccountNumber,
		&r.FirstName,
		&r.LastName,
		&r.Balance,
		&r.Email,
	)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	tran.Sen_acc = *s
	tran.Rec_acc = *r

	return tran, err

}
