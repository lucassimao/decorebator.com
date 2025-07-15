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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User
type UserRepository struct {
	Db *pgxpool.Pool
}

// Save creates a new user in the database
// country parameter can be nil, which will insert NULL into the database
// Note: Validation should be performed at the service layer before calling this function
func (repository *UserRepository) Save(ctx context.Context, firstName, lastName, password, email string, country *string) (*User, error) {
	// Note: country parameter is optional and can be nil
	// PostgreSQL will store NULL for nil pointer values
	query := `
		INSERT INTO users (first_name, last_name, password_hash, email, country, subscription_plan)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, profile_picture_url, country, date_of_birth, preferred_language,
			subscription_plan, subscription_status, stripe_customer_id, subscription_ends_at`

	// Create User struct to populate directly from database
	user := &User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), common.GetBcryptCost())
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(bytes)

	// Handle country pointer: dereference if non-nil, otherwise use nil for NULL
	var countryParam interface{}
	if country != nil {
		countryParam = *country
	} else {
		countryParam = nil
	}

	err = repository.Db.QueryRow(ctx, query, firstName, lastName, user.PasswordHash, email, countryParam, model.PlanFree).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.ProfilePictureURL, &user.Country, &user.DateOfBirth, &user.PreferredLanguage,
		&user.SubscriptionPlan, &user.SubscriptionStatus, &user.StripeCustomerID, &user.SubscriptionEndsAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				// Handle both old and new email constraint names for migration compatibility
				if strings.Contains(pgErr.ConstraintName, "users_email_unique") {
					return nil, common.BusinessError{Message: "Email already exists."}
				}
			}
		}
		return nil, err
	}

	return user, nil
}

type FindUserArgs struct {
	Email            *string
	ID               *int64
	StripeCustomerID *string
}

func (repository *UserRepository) Find(ctx context.Context, args FindUserArgs) ([]User, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT id, email, first_name, last_name, password_hash, 
		profile_picture_url, country, date_of_birth, preferred_language,
		subscription_plan, subscription_status, stripe_customer_id,
		platform, subscription_ends_at, created_at, updated_at FROM users`)
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
	}

	if len(whereConditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(whereConditions, " AND "))
	}

	users := []User{}
	query := builder.String()
	rows, err := repository.Db.Query(ctx, query, queryArgs...)
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
			&user.SubscriptionPlan, &user.SubscriptionStatus, &user.StripeCustomerID,
			&user.Platform, &user.SubscriptionEndsAt, &user.CreatedAt, &user.UpdatedAt)
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

func (repository *UserRepository) UpdatePassword(ctx context.Context, userID int64, newPassword string) error {
	query := `UPDATE users SET password_hash = $1, updated_at=NOW() WHERE ID = $2`

	bytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), common.GetBcryptCost())
	passwordHash := string(bytes)
	if err != nil {
		return err
	}

	_, err = repository.Db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return err
	}

	return nil
}

func (repository *UserRepository) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE ID = $1`
	_, err := repository.Db.Exec(ctx, query, userID)
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

func (repository *UserRepository) UpdateUserProfile(ctx context.Context, args UpdateUserProfileArgs) (*User, error) {
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
		bytes, err := bcrypt.GenerateFromPassword([]byte(*args.Password), common.GetBcryptCost())
		if err != nil {
			return nil, err
		}
		hash := string(bytes)
		passwordHash = &hash
	}

	user := User{}
	err := repository.Db.QueryRow(ctx, query,
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

// UpdateSubscriptionPlan updates a user's subscription plan
func (repository *UserRepository) UpdateSubscriptionPlan(ctx context.Context, userID int64, plan model.SubscriptionPlan) error {
	query := `UPDATE users SET subscription_plan = $1, updated_at = NOW() WHERE id = $2`
	_, err := repository.Db.Exec(ctx, query, plan, userID)
	if err != nil {
		return fmt.Errorf("failed to update subscription plan: %w", err)
	}
	return nil
}
