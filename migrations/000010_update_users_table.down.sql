-- remove the role_id column from the users table
ALTER TABLE users
DROP COLUMN IF EXISTS role_id;  