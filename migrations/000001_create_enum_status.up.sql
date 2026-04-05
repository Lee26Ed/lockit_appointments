-- Enums
CREATE TYPE user_status AS ENUM ('active', 'pending', 'suspended');

CREATE TYPE business_status AS ENUM ('active', 'inactive', 'suspended');

CREATE TYPE appointment_status AS ENUM (
  'pending',
  'confirmed',
  'cancelled',
  'completed',
  'no_show'
);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;