package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

type Service struct {
	ID          int     `json:"id"`
	BusinessID  int     `json:"business_id"`
	BusinessName string  `json:"business_name,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Duration    int     `json:"duration"` // in minutes
	Price       float64 `json:"price"`
	Active 	bool    `json:"active"`
}

type ServiceModel struct {
	DB *sql.DB
}

func ValidateService(v *validator.Validator, service *Service) {
	v.Check(service.Name != "", "name", "must be provided")
	v.Check(len(service.Name) <= 255, "name", "must not be more than 255 characters long")
	v.Check(service.Price >= 0, "price", "must be greater than or equal to zero")
	v.Check(service.Duration >= 0, "duration", "must be greater than or equal to zero")
}

func (s *ServiceModel) Insert(service *Service) error {
	query := `
	INSERT INTO services (business_id, name, description, duration_minutes, price, active)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.DB.QueryRowContext(ctx, query, service.BusinessID, service.Name, service.Description, service.Duration, service.Price, service.Active).Scan(&service.ID)
}


func (s *ServiceModel) GetAll(filters Filters) ([]*Service, Metadata, error) {
	query := `
		SELECT count(*) OVER() AS total_count, 
			s.id, 
			s.business_id, 
			b.name AS business_name, 
			s.name, 
			s.description, 
			s.duration_minutes, 
			s.price, 
			s.active
		FROM services s
		JOIN businesses b 
			ON s.business_id = b.id
		ORDER BY s.` + filters.sortColumn() + ` ` + filters.sortDirection() + `
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	services := []*Service{}

	for rows.Next() {
		var service Service

		err := rows.Scan(
			&totalRecords,
			&service.ID,
			&service.BusinessID,
			&service.BusinessName,
			&service.Name,
			&service.Description,
			&service.Duration,
			&service.Price,
			&service.Active,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		services = append(services, &service)

		if err = rows.Err(); err != nil {
			return nil, Metadata{}, err
		}
	}

		metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

		return services, metadata, nil
}