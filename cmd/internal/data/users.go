package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID int				`json:"id"`
	Email string		`json:"email"`
	Username string		`json:"username"`
	Password password	`json:"-"`
	RoleID int			`json:"role_id"`
	CreatedAt time.Time	`json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at,omitempty"`
}

type password struct {
	plaintext *string
	hash []byte
}

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

// Insert a new user into the database and return the ID of the newly created user
func (u *UserModel) Insert(user *User) (*User, error) {
	query := `
		INSERT INTO users (
			email, username, password_hash, role_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, created_at
	`

	args := []interface{}{
		user.Email,
		user.Username,
		user.Password.hash,
		2,
		time.Now(),
		time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

		err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		// detect duplicate email error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrDuplicateEmail
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
		SELECT id, username, email, password_hash, role_id, created_at
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
		&user.RoleID,
		&user.CreatedAt,
	)

	user.UpdatedAt = time.Now()
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

func (u *UserModel) GetAll() ([]*User, error) {
	query := `
		SELECT id, username, email, password_hash, role_id, created_at
		FROM users
		ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.RoleID,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

var ErrEditConflict = errors.New("edit conflict")

// updates a user record in the database
func (m *UserModel) Update(user *User) (*User, error) {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, role_id = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING username, email, role_id, updated_at
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
		&user.Username,
		&user.Email,
		&user.RoleID,
		&user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrDuplicateEmail
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEditConflict
		}
		return nil, err
	}

	return user, nil
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