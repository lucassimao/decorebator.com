ALTER TABLE users
ADD COLUMN profile_picture_url TEXT,
ADD COLUMN country VARCHAR(100),
ADD COLUMN date_of_birth DATE,
ADD COLUMN preferred_language VARCHAR(10) DEFAULT 'en';