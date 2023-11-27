package users

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64            `json:"id"`
	FirstName    string           `json:"firstName"`
	LastName     string           `json:"lastName"`
	PasswordHash string           `json:"passwordHash"`
	Email        string           `json:"email"`
	CreatedAt    pgtype.Timestamp `json:"createdAt"`
	UpdatedAt    pgtype.Timestamp `json:"updatedAt"`
}

func (u User) MarshalJSON() ([]byte, error) {
	userMap := map[string]interface{}{
		"id":           u.ID,
		"firstName":    u.FirstName,
		"lastName":     u.LastName,
		"passwordHash": u.PasswordHash,
		"email":        u.Email,
	}

	if u.CreatedAt.Status == pgtype.Present {
		userMap["createdAt"] = u.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if u.UpdatedAt.Status == pgtype.Present {
		userMap["updatedAt"] = u.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	return json.Marshal(userMap)
}

type UserRepository struct {
	db *pgxpool.Pool
}

func (repository *UserRepository) save(firstName, lastName, password, email string) (*User, error) {
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

	err = repository.db.QueryRow(context.Background(), query, firstName, lastName, passwordHash, email).Scan(&userID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &User{userID, firstName, lastName, passwordHash, email, createdAt, updatedAt}, nil
}

type FindArgs struct {
	email *string
}

func (repository *UserRepository) find(args FindArgs) ([]*User, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id,first_name, last_name, password_hash, created_at, updated_at FROM users`)
	var queryArgs []interface{}

	if args.email != nil {
		builder.WriteString(` WHERE email = $1`)
		queryArgs = append(queryArgs, args.email)
	}

	users := []*User{}
	query := builder.String()
	rows, err := repository.db.Query(context.Background(), query, queryArgs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		w := User{}
		err := rows.Scan(&w.ID, &w.FirstName, &w.LastName, &w.PasswordHash, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
