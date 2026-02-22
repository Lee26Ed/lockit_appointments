CREATE TABLE reviews (
  id SERIAL PRIMARY KEY,
  appointment_id INT NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
  business_id INT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
  rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  comment TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_reviews_business_id ON reviews(business_id);