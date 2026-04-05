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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
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