ALTER TABLE movies
ADD CONSTRAINT version_check
CHECK (version<=10);