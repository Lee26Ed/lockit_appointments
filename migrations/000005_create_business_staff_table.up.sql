CREATE TABLE business_staff (
  id SERIAL PRIMARY KEY,
  business_id INT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id INT NOT NULL REFERENCES roles(id),
  active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE (business_id, user_id)
);

CREATE INDEX idx_business_staff_business_id ON business_staff(business_id);
CREATE INDEX idx_business_staff_user_id ON business_staff(user_id);
CREATE INDEX idx_business_staff_role_id ON business_staff(role_id);