package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
		RETURNING id, created_at, updated_at, profile_picture_url, country, date_of_birth, preferred_language,
			subscription_plan, subscription_status, stripe_customer_id, subscription_ends_at`

	var userID int64
	var createdAt pgtype.Timestamp
	var updatedAt pgtype.Timestamp
	var profilePictureURL *string
	var country *string
	var dateOfBirth *time.Time
	var preferredLanguage *string
	var subscriptionPlan model.SubscriptionPlan
	var subscriptionStatus *model.SubscriptionStatus
	var stripeCustomerID *string
	var subscriptionEndsAt *time.Time

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	passwordHash := string(bytes)
	if err != nil {
		return nil, err
	}

	err = repository.Db.QueryRow(context.Background(), query, firstName, lastName, passwordHash, email).Scan(
		&userID, &createdAt, &updatedAt, &profilePictureURL, &country, &dateOfBirth, &preferredLanguage,
		&subscriptionPlan, &subscriptionStatus, &stripeCustomerID, &subscriptionEndsAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return nil, common.BusinessError{Message: "Email already exists."}
			}
		}
		return nil, err
	}

	return &User{
		ID: userID, FirstName: firstName, LastName: lastName, PasswordHash: passwordHash, Email: email,
		ProfilePictureURL: profilePictureURL, Country: country, DateOfBirth: dateOfBirth, PreferredLanguage: preferredLanguage,
		SubscriptionPlan: subscriptionPlan, SubscriptionStatus: subscriptionStatus,
		StripeCustomerID: stripeCustomerID, SubscriptionEndsAt: subscriptionEndsAt,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

type FindUserArgs struct {
	Email            *string
	ID               *int64
	StripeCustomerID *string
}

func (repository *UserRepository) Find(args FindUserArgs) ([]User, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id, email, first_name, last_name, password_hash, 
		profile_picture_url, country, date_of_birth, preferred_language,
		subscription_plan, subscription_status, stripe_customer_id, subscription_ends_at,
		created_at, updated_at FROM users`)
	var queryArgs []interface{}
	var whereConditions []string

	argIndex := 1
	if args.Email != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(email) = LOWER($%d)", argIndex))
		queryArgs = append(queryArgs, args.Email)
		argIndex++
	}

	if args.ID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("id = $%d", argIndex))
		queryArgs = append(queryArgs, args.ID)
		argIndex++
	}

	if args.StripeCustomerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("stripe_customer_id = $%d", argIndex))
		queryArgs = append(queryArgs, args.StripeCustomerID)
		argIndex++
	}

	if len(whereConditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(whereConditions, " AND "))
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
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash,
			&user.ProfilePictureURL, &user.Country, &user.DateOfBirth, &user.PreferredLanguage,
			&user.SubscriptionPlan, &user.SubscriptionStatus, &user.StripeCustomerID, &user.SubscriptionEndsAt,
			&user.CreatedAt, &user.UpdatedAt)
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

func (repository *UserRepository) Delete(userId int64) error {
	query := `DELETE FROM users WHERE ID = $1`
	_, err := repository.Db.Exec(context.Background(), query, userId)
	return err
}

type UpdateUserProfileArgs struct {
	ID                int64
	FirstName         *string
	LastName          *string
	Country           *string
	DateOfBirth       *time.Time
	PreferredLanguage *string
	ProfilePictureURL *string
	Password          *string
}

func (repository *UserRepository) UpdateUserProfile(args UpdateUserProfileArgs) (*User, error) {
	// The COALESCE defaults to the existing value if the new value is NULL
	query := `UPDATE users 
		SET first_name = COALESCE($2,first_name),
		    last_name = COALESCE($3,last_name), 
		    country = COALESCE($4,country), 
		    date_of_birth = COALESCE($5,date_of_birth), 
		    preferred_language = COALESCE($6,preferred_language),
		    updated_at = NOW(),
			profile_picture_url = COALESCE($7,profile_picture_url),
			password_hash = COALESCE($8,password_hash)
		WHERE id = $1
		RETURNING id, email, first_name, last_name, password_hash, 
			profile_picture_url, country, date_of_birth, preferred_language,
			subscription_plan, subscription_status, stripe_customer_id, subscription_ends_at,
			created_at, updated_at`

	var passwordHash *string
	if args.Password != nil {
		bytes, err := bcrypt.GenerateFromPassword([]byte(*args.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hash := string(bytes)
		passwordHash = &hash
	}

	user := User{}
	err := repository.Db.QueryRow(context.Background(), query,
		args.ID, args.FirstName, args.LastName, args.Country, args.DateOfBirth, args.PreferredLanguage,
		args.ProfilePictureURL, passwordHash).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash,
		&user.ProfilePictureURL, &user.Country, &user.DateOfBirth, &user.PreferredLanguage,
		&user.SubscriptionPlan, &user.SubscriptionStatus, &user.StripeCustomerID, &user.SubscriptionEndsAt,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.BusinessError{Message: "User not found"}
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByRevenueCatCustomerID retrieves a user by their RevenueCat customer ID
func (repository *UserRepository) GetUserByRevenueCatCustomerID(ctx context.Context, revenueCatCustomerID string) (*model.User, error) {
	query := `
		SELECT id, first_name, last_name, password_hash, email, profile_picture_url, 
		       country, date_of_birth, preferred_language, subscription_plan, 
		       subscription_status, stripe_customer_id, revenuecat_customer_id, 
		       platform, subscription_ends_at, created_at, updated_at
		FROM users 
		WHERE revenuecat_customer_id = $1
	`

	var user model.User
	err := repository.Db.QueryRow(ctx, query, revenueCatCustomerID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.PasswordHash, &user.Email,
		&user.ProfilePictureURL, &user.Country, &user.DateOfBirth, &user.PreferredLanguage,
		&user.SubscriptionPlan, &user.SubscriptionStatus, &user.StripeCustomerID,
		&user.RevenueCatCustomerID, &user.Platform, &user.SubscriptionEndsAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
