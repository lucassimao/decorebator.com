package repository

import (
	"context"
	"errors"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User
type UserRepository struct {
	Db *pgxpool.Pool
}

func (repository *UserRepository) Save(firstName, lastName, password, email string) (*User, error) {
	query := `
		INSERT INTO users (first_name, last_name, password_hash, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	var userID int64
	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	passwordHash := string(bytes)
	if err != nil {
		return nil, err
	}

	err = repository.Db.QueryRow(context.Background(), query, firstName, lastName, passwordHash, email).Scan(&userID, &createdAt, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return nil, common.BusinessError{Message: "Email already exists."}
			}
		}
		return nil, err
	}

	return &User{ID: userID, FirstName: firstName, LastName: lastName, PasswordHash: passwordHash, Email: email,
		CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

type FindUserArgs struct {
	Email *string
}

func (repository *UserRepository) Find(args FindUserArgs) ([]User, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id,email,first_name, last_name, password_hash, created_at, updated_at FROM users`)
	var queryArgs []interface{}

	if args.Email != nil {
		builder.WriteString(` WHERE LOWER(email) = LOWER($1)`)
		queryArgs = append(queryArgs, args.Email)
	}

	users := []User{}
	query := builder.String()
	rows, err := repository.Db.Query(context.Background(), query, queryArgs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (repository *UserRepository) UpdatePassword(userId int64, newPassword string) error {
	query := `UPDATE users SET password_hash = $1, updated_at=NOW() WHERE ID = $2`

	bytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	passwordHash := string(bytes)
	if err != nil {
		return err
	}

	_, err = repository.Db.Exec(context.Background(), query, passwordHash, userId)
	if err != nil {
		return err
	}

	return nil
}
