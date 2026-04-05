CREATE TABLE businesses (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  bio VARCHAR(500),

  owner_id INT REFERENCES users(id),

  email VARCHAR(255),
  phone VARCHAR(25),

  logo_url VARCHAR(255),
  slug VARCHAR(255) UNIQUE NOT NULL,

  status business_status,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP
);

CREATE TRIGGER set_businesses_updated_at
BEFORE UPDATE ON businesses
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();