package data

import "database/sql"


type Models struct {
	Users *UserModel
	Services *ServiceModel
	Businesses *BusinessModel
}

func CreateModels(db *sql.DB) *Models {
	return &Models{
		Users: &UserModel{DB: db},
		Services: &ServiceModel{DB: db},
		Businesses: &BusinessModel{DB: db},
	}
}