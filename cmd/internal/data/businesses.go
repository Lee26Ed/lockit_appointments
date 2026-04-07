package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

type Business struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Bio   string `json:"bio,omitempty"`
	OwnerID int   `json:"owner_id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	LogoURL string `json:"logo_url,omitempty"`
	Slug string `json:"slug"`
	Status BusinessStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type BusinessModel struct {
	DB *sql.DB
}

type BusinessStatus string

const (
	BusinessStatusActive    BusinessStatus = "active"
	BusinessStatusInactive   BusinessStatus = "inactive"
	BusinessStatusSuspended BusinessStatus = "suspended"
)

func ValidateBusiness(v *validator.Validator, business *Business) {
	v.Check(business.Name != "", "name", "must be provided")
	v.Check(len(business.Name) <= 255, "name", "must not be more than 255 characters long")

	v.Check(len(business.Bio) <= 500, "bio", "must not be more than 500 characters long")

	ValidateEmail(v, business.Email)

	v.Check(business.OwnerID != 0, "owner_id", "must not be empty")
}

var ErrDuplicateSlug = errors.New("duplicate slug")
var ErrOwnerIDInvalid = errors.New("owner_id does not reference a valid user")

func (b *BusinessModel) SlugExists(slug string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM businesses
			WHERE slug = $1
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	err := b.DB.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (b *BusinessModel) GenerateUniqueSlug(name string) (string, error) {
	baseSlug := utils.GenerateSlug(name)
	if baseSlug == "" {
		return "", nil
	}

	slug := baseSlug
	for suffix := 2; ; suffix++ {
		exists, err := b.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
}

func (b *BusinessModel) Insert(business *Business) (*Business, error) {
	query := `
		INSERT INTO businesses (name, bio, owner_id, email, phone, logo_url, slug, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	args := []interface{}{
		business.Name,
		business.Bio,
		business.OwnerID,
		business.Email,
		business.Phone,
		business.LogoURL,
		business.Slug,
		business.Status,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, args...).Scan(&business.ID, &business.CreatedAt)
	if err != nil {
		// detect duplicate slug error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && strings.Contains(err.Error(), "businesses_slug_key") {
			return nil, ErrDuplicateSlug
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") && strings.Contains(err.Error(), "businesses_owner_id_fkey") {
			return nil, ErrOwnerIDInvalid
		}
		return nil, err
	}

	return business, nil
}

func (b *BusinessModel) GetAll(filters Filters) ([]*Business, Metadata, error) {
	query := `
		SELECT count(*) OVER() AS total_count,
		id, 
		name,
		bio,
		owner_id,
		email,
		phone,
		logo_url,
		slug,
		status,
		created_at,
		updated_at
		FROM businesses
		ORDER BY ` + filters.sortColumn() + ` ` + filters.sortDirection() + `
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := b.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	businesses := []*Business{}
	totalRecords := 0

	for rows.Next() {
		var business Business

		err := rows.Scan(
			&totalRecords,
			&business.ID,
			&business.Name,
			&business.Bio,
			&business.OwnerID,
			&business.Email,
			&business.Phone,
			&business.LogoURL,
			&business.Slug,
			&business.Status,
			&business.CreatedAt,
			&business.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		businesses = append(businesses, &business)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return businesses, metadata, nil
}

func (b *BusinessModel) Get(id int) (*Business, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, bio, owner_id, email, phone, logo_url, slug, status, created_at, updated_at
		FROM businesses
		WHERE id = $1`

	var business Business

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(
		&business.ID,
		&business.Name,
		&business.Bio,
		&business.OwnerID,
		&business.Email,
		&business.Phone,
		&business.LogoURL,
		&business.Slug,
		&business.Status,
		&business.CreatedAt,
		&business.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &business, nil
}

func (b *BusinessModel) GetByOwnerID(ownerID int) (*Business, error) {
	if ownerID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, bio, owner_id, email, phone, logo_url, slug, status, created_at, updated_at
		FROM businesses
		WHERE owner_id = $1`

	var business Business

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, ownerID).Scan(
		&business.ID,
		&business.Name,
		&business.Bio,
		&business.OwnerID,
		&business.Email,
		&business.Phone,
		&business.LogoURL,
		&business.Slug,
		&business.Status,
		&business.CreatedAt,
		&business.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &business, nil
}

func (b *BusinessModel) Update(business *Business) error {
	query := `
		UPDATE businesses
		SET name = $1, bio = $2, owner_id = $3, email = $4, phone = $5, logo_url = $6, slug = $7, status = $8
		WHERE id = $9
	`

	args := []interface{}{
		business.Name,
		business.Bio,
		business.OwnerID,
		business.Email,
		business.Phone,
		business.LogoURL,
		business.Slug,
		business.Status,
		business.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.DB.ExecContext(ctx, query, args...)
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

func (b *BusinessModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM businesses
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.DB.ExecContext(ctx, query, id)
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