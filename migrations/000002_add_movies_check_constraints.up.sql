-- Movie runtime must be positive
ALTER TABLE movies ADD CONSTRAINT movies_runtime_check CHECK (runtime >= 0);

-- Movie year must be between 1888 and current year
ALTER TABLE movies ADD CONSTRAINT movies_year_check CHECK (year BETWEEN 1888 AND date_part('year', now()));

-- There must be at least 1 genre and less than equal to 5 genres
ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK (array_length(genres, 1) BETWEEN 1 AND 5);
