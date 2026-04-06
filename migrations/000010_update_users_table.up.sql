-- alter the users table to add a role_id column that references the roles table
ALTER TABLE users
ADD COLUMN role_id INT REFERENCES roles(id) ON DELETE SET NULL;
