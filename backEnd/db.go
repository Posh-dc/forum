package backEnd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStoredData struct {
	UserID        string    `json:"id"`
	UserName      string    `json:"username"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	Password      string    `json:"password"`
	PhoneNumber   string    `json:"phone_number"`
	CreatedAT     time.Time `json:"created_at"`
	UpdatedAT     time.Time `json:"updated_at"`
	EmailVerified bool      `json:"is_email_verified"`
}

type UserEmailVerificationCodeStoredData struct {
	ID        int64      `json:"id"`
	UserID    string     `json:"user_id"`
	CodeHash  string     `json:"code_hash"`
	Attempts int32      `json:"attempts"`
	CreatedAT time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAT    *time.Time `json:"used_at"`
}

type UserSessionStoredData struct {
	SessionValue string    `json:"id"`
	UserID       string    `json:"user_id"`
	CreatedAT    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// CHECKING IF EMAIL, USERNAME AND  PHONE NUMBER ALREADY EXIST
func CheckAlreadyExist(column string, value any, pool *pgxpool.Pool, ctx context.Context) (bool, error) {

	query := fmt.Sprintf("SELECT EXISTS (SELECT * from users where %v = $1)", column)

	var exist bool
	err := pool.QueryRow(ctx, query, value).Scan(&exist)

	if err != nil {
		return false, err
	}

	return exist, nil
}

// FUNCION TO INSER USER REGISTRATION DETAILS STARTS HERE
func InsertUser(user UserRegInfo, hashedSessionID string, pool *pgxpool.Pool, ctx context.Context) error {

	tx, err := pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	var userID string

	//             insert user details
	err = tx.QueryRow(ctx,
		`INSERT INTO users (username,display_name,email,password,phone_number,created_at )
	  VALUES ($1, $2, $3, $4, $5, $6) returning id;`,
		user.UserName,
		user.DisplayName,
		user.Email,
		user.Password,
		user.PhoneNumber,
		time.Now(),
	).Scan(&userID)

	if err != nil {
		return err
	}

	now := time.Now()

	//             insert user session-id
	_, err = tx.Exec(ctx, `INSERT INTO session(id, user_id, expires_at) VALUES($1,$2,$3);`,
		hashedSessionID,
		userID,
		now.Add(time.Hour))

	if err != nil {
		return err
	}

	// generating and saving verification code

	code, err := GenerateVerificationCode()

	if err != nil {
		return fmt.Errorf("Generating Verification code Error: %w", err)
	}

	codeHashed := GeneralHashFunction(code)

	// inserting verification code
	_, err = tx.Exec(ctx, `INSERT INTO email_verification_codes(user_id, code_hash, expires_at) VALUES($1,$2,$3)`,
		userID,
		codeHashed,
		now.Add(10*time.Minute),
	)

	if err != nil {
		return err
	}

	// email_outbox starts here

	_, err = tx.Exec(ctx, `INSERT INTO email_outbox(recipient_email, subject, body) VALUES($1,$2,$3)`,
		user.Email,
		EmailVerificationSubject,
		CreateVerificationEmailBody(code),
	)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil

}

// FUNCTION TO CHECK UNIQUENESS
func CheckUniqueConstraint(insertingError error) (error, int, string) {
	var pgErr *pgconn.PgError

	if errors.As(insertingError, &pgErr) {
		if pgErr.ConstraintName == "users_phone_number_key" {
			return fmt.Errorf("Phone Number Already Exist"), http.StatusConflict, "phoneNumber"
		} else if pgErr.ConstraintName == "users_email_key" {
			return fmt.Errorf("Email Already Exist"), http.StatusConflict, "email"
		} else if pgErr.ConstraintName == "users_username_key" {
			return fmt.Errorf("Username Already Exist"), http.StatusConflict, "userName"
		} else {
			return insertingError, http.StatusInternalServerError, ""
		}

	} else {
		return insertingError, http.StatusInternalServerError, ""
	}
}

// GETTING USER SESSION DETAILS
func GetUserSessionDetails(ctx context.Context, pool *pgxpool.Pool, sessionValue string) (UserSessionStoredData, error) {

	var data UserSessionStoredData

	err := pool.QueryRow(ctx, `SELECT * FROM session WHERE id = $1`, sessionValue).Scan(
		&data.SessionValue,
		&data.UserID,
		&data.CreatedAT,
		&data.ExpiresAt,
	)

	if err != nil {
		return data, err
	}

	return data, nil
}

// GETTING USER USING SESSION ID
func GetUserDetails(ctx context.Context, pool *pgxpool.Pool, userID string) (UserStoredData, error) {

	var data UserStoredData

	err := pool.QueryRow(ctx, `SELECT * FROM users WHERE id = $1`, userID).Scan(
		&data.UserID,
		&data.UserName,
		&data.DisplayName,
		&data.Email,
		&data.Password,
		&data.PhoneNumber,
		&data.CreatedAT,
		&data.UpdatedAT,
		&data.EmailVerified,
	)

	if err != nil {
		return data, err
	}

	return data, nil
}

// GETTING USER EMAIL VERIFICATION DETAILS
func GetUserEmailVerificationDetails(ctx context.Context, pool *pgxpool.Pool, userID string) (UserEmailVerificationCodeStoredData, error) {

	var data UserEmailVerificationCodeStoredData

	err := pool.QueryRow(ctx, `SELECT * FROM email_verification_codes WHERE user_id = $1`, userID).Scan(
		&data.ID,
		&data.UserID,
		&data.CodeHash,
		&data.Attempts,
		&data.CreatedAT,
		&data.ExpiresAt,
		&data.UsedAT,
	)

	if err != nil {
		return data, err
	}

	return data, nil

}
