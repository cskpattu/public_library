package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"public_library/utils"
	"strings"
)

var ErrNotFound = errors.New("book not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAllBooks(ctx context.Context, req PaginationRequest) ([]BookResponse, int64, int64, error) {
	log.Printf("<--------ListAllBooks starts-------->")
	defer log.Printf("<--------ListAllBooks ends-------->")

	var (
		responses  []BookResponse
		totalCount int64
	)

	// Build WHERE clauses and args
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "1=1") // base condition

	if req.Search != "" {
		whereClauses = append(whereClauses, "currency ILIKE ?")
		args = append(args, "%"+req.Search+"%")
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Pagination
	limit := req.PageSize
	if limit == 0 {
		limit = 10
	}
	offset := (req.Page - 1) * limit

	// --- Count Query ---
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, utils.BooksTable, whereSQL)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		log.Printf("Failed to count books: %v", err)
		return nil, 0, 0, err
	}

	// --- Data Query ---
	dataQuery := fmt.Sprintf(`
	SELECT
		id,
		title,
		author,
		isbn
	FROM %s
	WHERE %s
	LIMIT $%d OFFSET $%d
`, utils.BooksTable, whereSQL, len(args)+1, len(args)+2)

	argsWithPagination := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, argsWithPagination...)
	if err != nil {
		log.Printf("Failed to fetch books: %v", err)
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var b BookResponse
		err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.Author,
			&b.ISBN,
		)
		if err != nil {
			log.Printf("Failed to scan book row: %v", err)
			return nil, 0, 0, err
		}
		responses = append(responses, b)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return nil, 0, 0, err
	}

	return responses, int64(len(responses)), totalCount, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*Book, error) {
	log.Println("<--------GetByID starts-------->")
	defer log.Println("<--------GetByID ends-------->")

	const query = `
		SELECT id, title, author, isbn
		FROM books
		WHERE id = $1
	`

	var b Book
	err := r.db.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.Title, &b.Author, &b.ISBN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Book with id=%d not found", id)
			return nil, ErrNotFound
		}
		log.Printf("Failed to get book by id=%d: %v", id, err)
		return nil, err
	}

	return &b, nil
}

func (r *Repository) Create(ctx context.Context, b *Book) error {
	log.Println("<--------Create starts-------->")
	defer log.Println("<--------Create ends-------->")

	const query = `
		INSERT INTO books (title, author, isbn)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, b.Title, b.Author, b.ISBN).Scan(&b.ID)
	if err != nil {
		log.Printf("Failed to create book %+v: %v", b, err)
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, b *Book) error {
	log.Println("<--------Update starts-------->")
	defer log.Println("<--------Update ends-------->")

	const query = `
		UPDATE books
		SET title = $1, author = $2, isbn = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, b.Title, b.Author, b.ISBN, b.ID)
	if err != nil {
		log.Printf("Failed to update book id=%d: %v", b.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected for book id=%d update: %v", b.ID, err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No book found to update with id=%d", b.ID)
		return ErrNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	log.Println("<--------Delete starts-------->")
	defer log.Println("<--------Delete ends-------->")

	const query = `
		DELETE FROM books WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Failed to delete book id=%d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected for book id=%d delete: %v", id, err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No book found to delete with id=%d", id)
		return ErrNotFound
	}

	return nil
}
