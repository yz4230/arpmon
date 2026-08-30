ALTER TABLE devices
ALTER COLUMN mac_address TYPE VARCHAR(17)
USING BTRIM(mac_address);
