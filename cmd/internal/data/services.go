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
	DownTime    int     `json:"downtime"` // in minutes
	Price       float64 `json:"price"`
	Active 	bool    `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
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

func (s *ServiceModel) Insert(service *Service) (*Service, error) {
	query := `
	INSERT INTO services (business_id, name, description, duration_mins, downtime_mins, price, active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at`

	args := []interface{}{
		service.BusinessID,
		service.Name,
		service.Description,
		service.Duration,
		service.DownTime,
		service.Price,
		service.Active,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&service.ID, &service.CreatedAt)
	if err != nil {
		return nil, err
	}
	return service, nil

}


func (s *ServiceModel) GetAll(filters Filters) ([]*Service, Metadata, error) {
	query := `
		SELECT count(*) OVER() AS total_count, 
			s.id, 
			s.business_id, 
			b.name AS business_name, 
			s.name, 
			s.description, 
			s.duration_mins, 
			s.downtime_mins, 
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
			&service.DownTime,
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

func (s *ServiceModel) Get(id int) (*Service, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT 
		s.id, 
		s.business_id, 
		b.name AS business_name, 
		s.name, 
		s.description, 
		s.duration_mins, 
		s.downtime_mins, 
		s.price, 
		s.active, 
		s.created_at, 
		s.updated_at
	FROM services s
	JOIN businesses b ON s.business_id = b.id
	WHERE s.id = $1`

	var service Service
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&service.ID,
		&service.BusinessID,
		&service.BusinessName,
		&service.Name,
		&service.Description,
		&service.Duration,
		&service.DownTime,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	
	return &service, nil
}

func (s *ServiceModel) Update(service *Service) error {
	query := `
	UPDATE services
	SET name = $1, description = $2, duration_mins = $3, downtime_mins = $4, price = $5, active = $6
	WHERE id = $7`

	args := []interface{}{
		service.Name,
		service.Description,
		service.Duration,
		service.DownTime,
		service.Price,
		service.Active,
		service.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := s.DB.ExecContext(ctx, query, args...)
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

func (s *ServiceModel) Delete(id int) error {
	query := `DELETE FROM services WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := s.DB.ExecContext(ctx, query, id)
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