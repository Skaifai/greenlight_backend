-- make sure that the runtime value is always greater than zero
ALTER TABLE movies 
        ADD CONSTRAINT movies_runtime_check 
        CHECK (runtime>=0);
        
-- year value is between 1888 and the current year
ALTER TABLE movies 
        ADD CONSTRAINT movies_year_check 
        CHECK (year BETWEEN 1888 AND date_part('year',now()));

-- genres array always contains between 1 and 5 items.
ALTER TABLE movies 
        ADD CONSTRAINT genres_length_check 
        CHECK (array_length(genres, 1) BETWEEN 1 AND 5);