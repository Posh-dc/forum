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
func InsertUser(user UserRegInfo, sessionID string, pool *pgxpool.Pool, ctx context.Context) error {

	tx, err := pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	var id string

	err = tx.QueryRow(ctx,
		`INSERT INTO users (username,display_name,email,password,phone_number,created_at )
	  VALUES ($1, $2, $3, $4, $5, $6) returning id;`,
		user.UserName,
		user.DisplayName,
		user.Email,
		user.Password,
		user.PhoneNumber,
		time.Now(),
	).Scan(&id)

	if err != nil {
		return err
	}

	now := time.Now()

	tag, err := tx.Exec(ctx, `INSERT INTO session(id, user_id, expires_at) VALUES($1,$2,$3);`,
		sessionID,
		id,
		now.Add(time.Hour))

	if err != nil {
		return err
	}

	fmt.Printf("Rows Affected %v\n ", tag.RowsAffected())

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil

}

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
