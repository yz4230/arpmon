CREATE TABLE members(
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE devices(
  id SERIAL PRIMARY KEY,
  mac_address VARCHAR(17) NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scan_runs(
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  target_network CIDR NOT NULL
);

CREATE TABLE member_attendances(
  id SERIAL PRIMARY KEY,
  member_id INT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  scan_run_id INT NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
  is_present BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(scan_run_id, member_id)
);

CREATE TABLE device_observations(
  id SERIAL PRIMARY KEY,
  device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  scan_run_id INT NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
  ip_address INET NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(scan_run_id, device_id)
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_members_updated_at
BEFORE UPDATE ON members
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_member_attendances_updated_at
BEFORE UPDATE ON member_attendances
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_devices_updated_at
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scan_runs_updated_at
BEFORE UPDATE ON scan_runs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_device_observations_updated_at
BEFORE UPDATE ON device_observations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
