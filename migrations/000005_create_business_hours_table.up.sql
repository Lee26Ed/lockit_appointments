CREATE TABLE business_hours (
  id SERIAL PRIMARY KEY,
  business_id INT REFERENCES businesses(id),

  day_of_week INT NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,

  closed_for_lunch BOOLEAN DEFAULT FALSE,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP
);

CREATE TRIGGER set_business_hours_updated_at
BEFORE UPDATE ON business_hours
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();