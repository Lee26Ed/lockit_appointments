CREATE TABLE business_time_off (
  id SERIAL PRIMARY KEY,
  business_id INT REFERENCES businesses(id),

  start_datetime TIMESTAMP,
  end_datetime TIMESTAMP,

  reason VARCHAR(500)
);