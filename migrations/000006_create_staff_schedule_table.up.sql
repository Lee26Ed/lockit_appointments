CREATE TABLE staff_schedule (
  id SERIAL PRIMARY KEY,
  staff_id INT NOT NULL REFERENCES business_staff(id) ON DELETE CASCADE,
  day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  appointment_intervals INT,
  downtime_between_appointments INT
);

CREATE INDEX idx_staff_schedule_staff_id ON staff_schedule(staff_id);