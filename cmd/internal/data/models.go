package data

import "database/sql"


type Models struct {
	Users *UserModel
	Services *ServiceModel
	Businesses *BusinessModel
	Tokens *TokenModel
	Roles *RoleModel
}

func CreateModels(db *sql.DB) *Models {
	return &Models{
		Users: &UserModel{DB: db},
		Services: &ServiceModel{DB: db},
		Businesses: &BusinessModel{DB: db},
		Tokens: &TokenModel{DB: db},
		Roles: &RoleModel{DB: db},
	}
}