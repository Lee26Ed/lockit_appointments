CREATE TABLE services (
  id SERIAL PRIMARY KEY,
  business_id INT REFERENCES businesses(id),

  name VARCHAR(255),
  description VARCHAR(500),

  duration_mins INT,
  downtime_mins INT,

  price DECIMAL(10, 2),
  active BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP
);

CREATE TRIGGER set_services_updated_at
BEFORE UPDATE ON services
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();