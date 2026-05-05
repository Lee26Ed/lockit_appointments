package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var AnonymousUser = &User{}
type User struct {
	ID           int        `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Password     password     `json:"-"`
	Status       Status     `json:"status,omitempty"`
	IsActivated  bool       `json:"is_activated"`
	RoleID       int        `json:"role_id"`
	RoleName     string     `json:"role_name,omitempty"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type password struct {
	plaintext *string
	hash []byte
}

type Status string

const (
	UserStatusActive    Status = "active"
	UserStatusPending   Status = "pending"
	UserStatusSuspended Status = "suspended"
)

type UserModel struct {
	DB *sql.DB
}

func (p *password) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintext
	p.hash = hash
	return nil
}

// Compare the client-provided plaintext password with saved-hashed version
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// Validate the user data
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Username != "", "username", "must be provided")
	v.Check(len(user.Username) <= 100, "username", "must not be more than 100 bytes long")
	// validate email for user
	ValidateEmail(v, user.Email)
	// validate the plain text password
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// check if we messed up the password hash
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Validate the email address
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) <= 255, "email", "must not be more than 255 bytes long")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// Check that a valid password is provided
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

var ErrDuplicateEmail = errors.New("duplicate email")
var ErrDuplicateUsername = errors.New("duplicate username")

// Insert a new user into the database and return the ID of the newly created user
func (u *UserModel) Insert(user *User) (*User, error) {
	query := `
		INSERT INTO users (
			email, username, password_hash, status, is_activated, last_login, role_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id, created_at
	`

	args := []interface{}{
		user.Email,
		user.Username,
		user.Password.hash,
		user.Status,
		user.IsActivated,
		time.Now(),
		user.RoleID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		// detect duplicate email error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrDuplicateEmail
		}
		// detect duplicate username error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_username_key") {
			return nil, ErrDuplicateUsername
		}
		return nil, err
	}
	return user, nil
}

var ErrRecordNotFound = errors.New("record not found")

// Get a specific user based on its ID
func (u *UserModel) Get(id int) (*User, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, username, email, password_hash, status, is_activated, last_login, created_at, updated_at, role_id
		FROM users
		WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Status,
		&user.IsActivated,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.RoleID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserModel) GetAll(filters Filters) ([]*User, Metadata, error) {
	query := `
		SELECT count(*) OVER() as total_count,
		id, 
		username, 
		email, 
		password_hash, 
		status, 
		is_activated, 
		last_login, 
		created_at, 
		updated_at, 
		role_id
		FROM users
		ORDER BY ` + filters.sortColumn() + ` ` + filters.sortDirection() + `
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := u.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []*User{}
	
	for rows.Next() {
		var user User
		err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.Status,
			&user.IsActivated,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.RoleID,

		)
		if err != nil {
			return nil, Metadata{}, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return users, metadata, nil
}

// Get a user from the database based on their username provided
func (u *UserModel) GetByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role_id, status, is_activated, last_login, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.RoleID,
		&user.Status,
		&user.IsActivated,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u *UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.id, users.created_at, users.username,
               users.email, users.status, users.is_activated,
               users.role_id, roles.role
        FROM users
        INNER JOIN auth_tokens as tokens
        ON users.id = tokens.user_id
        INNER JOIN roles
        ON users.role_id = roles.id
        WHERE tokens.token = $1
        AND tokens.scope = $2 
        AND tokens.expires_at > $3
       `
	args := []any{tokenHash[:], tokenScope, time.Now()}
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := u.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.IsActivated,
		&user.RoleID,
		&user.RoleName,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Return the matching user.
	return &user, nil
}

func (u *UserModel) CanAccessUserData(currentUser *User, targetUserID int) (bool, error) {
		
	// Administrators can access any user
	if currentUser.RoleName == "admin" {
		return true, nil
	}
	
	// System Users can only access their own data
	return currentUser.ID == targetUserID, nil

}

var ErrEditConflict = errors.New("edit conflict")

// updates a user record in the database
func (m *UserModel) Update(user *User) (*User, error) {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, role_id = $4
		WHERE id = $5
		RETURNING id, username, email, updated_at
	`

	args := []interface{}{
		user.Username,
		user.Email,
		user.Password.hash,
		user.RoleID,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrDuplicateEmail
		}
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_username_key") {
			return nil, ErrDuplicateUsername
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEditConflict
		}
		return nil, err
	}

	return user, nil
}

// UpdateActivation updates only the is_active field for a user
// This is used when activating a user account via email token
func (m *UserModel) UpdateActivation(userID int, status Status, isActivated bool) error {
	query := `
		UPDATE users
		SET status = $1, is_activated = $2
		WHERE id = $3
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var updatedAt time.Time
	err := m.DB.QueryRowContext(ctx, query, status, isActivated, userID).Scan(&updatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

// Delete removes a user record from the database
func (u *UserModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := u.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// CountUsers returns the total number of users in the database
func (u *UserModel) CountUsers() (int, error) {
	query := `SELECT COUNT(*) FROM users`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := u.DB.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
