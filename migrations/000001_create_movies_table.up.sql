-- null values are not appreciated in GoLang 
-- So all columns either not null or have default vals
CREATE TABLE IF NOT EXISTS movies (
    -- id column is a 64-bit auto-incrementing integer & primary key (defines the row)
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone not null default NOW(), 
    title text not null,
    year integer not null,
    runtime integer not null,
    -- genres column is array of zero-or-more text values. 
    genres text[] not NULL,
    version integer NOT NULL DEFAULT 1
);

