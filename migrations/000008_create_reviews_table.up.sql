CREATE TABLE reviews (
  id SERIAL PRIMARY KEY,

  appointment_id INT UNIQUE REFERENCES appointments(id),
  business_id INT REFERENCES businesses(id),

  rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  comment TEXT,

  created_at TIMESTAMP DEFAULT NOW()
);