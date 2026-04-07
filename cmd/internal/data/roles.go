// Filename: internal/data/roles.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

// Role struct represents a system role
type Role struct {
	ID       int    `json:"id"`
	RoleName string `json:"role_name"`
}

// ValidateRole validates a role struct
func ValidateRole(v *validator.Validator, role *Role) {
	v.Check(role.RoleName != "", "role_name", "must be provided")
	v.Check(len(role.RoleName) <= 50, "role_name", "must not be more than 50 characters long")
}

// RoleModel wraps a database connection pool
type RoleModel struct {
	DB *sql.DB
}

// Insert a new role record in the database
func (r *RoleModel) Insert(role *Role) error {
	query := `
		INSERT INTO roles (role)
		VALUES ($1)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, role.RoleName).Scan(&role.ID)
}

// Get retrieves a specific role based on its ID
func (r *RoleModel) Get(id int) (*Role, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, role
		FROM roles
		WHERE id = $1`

	var role Role

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.RoleName,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}

// GetByName retrieves a role by its name (useful for authentication)
func (r *RoleModel) GetByName(name string) (*Role, error) {
	query := `
		SELECT id, role
		FROM roles
		WHERE role = $1`

	var role Role

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.RoleName,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}

// GetAll retrieves all roles from the database
func (r *RoleModel) GetAll(filters Filters) ([]*Role, Metadata, error) {
	query := `
		SELECT count(*) OVER() AS total_count,
		id, 
		role
		FROM roles
		ORDER BY ` + filters.sortColumn() + ` ` + filters.sortDirection() + `
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	roles := []*Role{}
	totalRecords := 0

	for rows.Next() {
		var role Role

		err := rows.Scan(
			&totalRecords,
			&role.ID,
			&role.RoleName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		roles = append(roles, &role)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return roles, metadata, nil
}

// Update an existing role record in the database
func (r *RoleModel) Update(role *Role) error {
	query := `
		UPDATE roles
		SET role = $2
		WHERE id = $1`

	args := []interface{}{
		role.ID,
		role.RoleName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

// Delete removes a role record from the database
func (r *RoleModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM roles WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, id)
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
