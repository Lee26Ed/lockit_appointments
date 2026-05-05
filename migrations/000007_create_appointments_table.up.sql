CREATE TABLE appointments (
  id SERIAL PRIMARY KEY,

  business_id INT REFERENCES businesses(id),
  service_id INT REFERENCES services(id),
  customer_id INT REFERENCES users(id),

  name VARCHAR(200),
  notes VARCHAR(500),

  start_time TIMESTAMP,
  end_time TIMESTAMP,

  status appointment_status DEFAULT 'pending',

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP
);

CREATE TRIGGER set_appointments_updated_at
BEFORE UPDATE ON appointments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();