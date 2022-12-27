package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Movie By default, the keys in the JSON object are equal to the field names in the struct ( ID,
// CreatedAt, Title and so on).
type Movie struct {
	ID        int64     `json:"id"`                       // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                        // Timestamp for when the movie is added to our database, "-" directive, hidden in response
	Title     string    `json:"title"`                    // Movie title
	Year      int32     `json:"year,omitempty"`           // Movie release year, "omitempty" - hide from response if empty
	Runtime   int32     `json:"runtime,omitempty,string"` // Movie runtime (in minutes), "string" - convert int to string
	Genres    []string  `json:"genres,omitempty"`         // Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     `json:"version"`                  // The version number starts at 1 and will be incremented each
	// time the movie information is updated
}

// MovieModel is a struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// Insert method for inserting a new record in the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies(title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no movies will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie
	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Otherwise, return a pointer to the Movie struct.
	return &movie, nil
}

// Update method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// Delete method for deleting a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	query := `
		DELETE FROM movies
		WHERE id= $1
		`
	// Error handling
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return nil
	}

	// Checking how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Check if the row was in the database before the query
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
