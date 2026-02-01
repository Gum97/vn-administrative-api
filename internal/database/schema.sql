-- Create provinces table
CREATE TABLE IF NOT EXISTS provinces (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create admin_units table (Wards/Communes linked to Province)
CREATE TABLE IF NOT EXISTS admin_units (
    id INT PRIMARY KEY,
    province_id INT NOT NULL,
    name TEXT NOT NULL,
    level TEXT,
    code TEXT,
    pre_merger_desc TEXT,
    lat FLOAT,
    long FLOAT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (province_id) REFERENCES provinces(id)
);

CREATE INDEX IF NOT EXISTS idx_admin_units_province ON admin_units(province_id);
CREATE INDEX IF NOT EXISTS idx_admin_units_name ON admin_units(name);
CREATE INDEX IF NOT EXISTS idx_admin_units_pre_merger ON admin_units(pre_merger_desc);

