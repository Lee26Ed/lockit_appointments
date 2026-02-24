package data

import "database/sql"


type Models struct {
	Users *UserModel
}

func CreateModels(db *sql.DB) *Models {
	return &Models{
		Users: &UserModel{DB: db},
	}
}