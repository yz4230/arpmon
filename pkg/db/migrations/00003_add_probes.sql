CREATE TABLE probes(
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE device_probes
ADD COLUMN probe_id INT REFERENCES probes(id) ON DELETE CASCADE;

INSERT INTO probes (created_at, updated_at)
SELECT created_at, MAX(updated_at)
FROM device_probes
GROUP BY created_at;

UPDATE device_probes
SET probe_id = probes.id
FROM probes
WHERE device_probes.created_at = probes.created_at;

ALTER TABLE device_probes
ALTER COLUMN probe_id SET NOT NULL;

CREATE TRIGGER update_probes_updated_at
BEFORE UPDATE ON probes
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
