ALTER TABLE services DROP CONSTRAINT IF EXISTS services_business_id_fkey;

ALTER TABLE services
ADD CONSTRAINT services_business_id_fkey
FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;
