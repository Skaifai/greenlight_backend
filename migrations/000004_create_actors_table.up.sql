-- null values are not appreciated in GoLang
-- So all columns either not null or have default vals
CREATE TABLE IF NOT EXISTS actors (
    -- id column is a 64-bit auto-incrementing integer & primary key (defines the row)
                                      id bigserial PRIMARY KEY,
                                      name VARCHAR,
                                      surname VARCHAR
);
