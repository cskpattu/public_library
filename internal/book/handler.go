package book

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"public_library/internal/db"
	"public_library/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	repo   *Repository
	logger *zap.Logger
	config db.AppConfig
}

func NewHandler(r *Repository, l *zap.Logger) *Handler {
	return &Handler{repo: r, logger: l}
}

// HealthCheck handles GET /health
// Always returns 200 OK. Status can be "ok" or "degraded"
// @Summary     Health check
// @Description Returns server health status
// @Tags        Health
// @Accept      */*
// @Produce     json
// @Success     200 {object} StatusResponse
// @Failure     200 {object} StatusResponse
// @Router      /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	timestamp := time.Now().UTC().Format(time.RFC3339)
	// Default health response
	response := StatusResponse{
		Status:    utils.StatusOK,
		Version:   "v1.0.0",
		Timestamp: timestamp,
		Message:   utils.StatusOK,
	}

	// Check the database connection
	if err := h.repo.db.PingContext(ctx); err != nil {
		// Log DB ping failure and return degraded status
		h.logger.Error("Health check: DB ping failed", zap.Error(err))
		response.Status = utils.StatusDegraded
		response.Message = utils.StatusError
		w.WriteHeader(http.StatusOK) // Return 200 OK for degraded status
	} else {
		// Log success
		h.logger.Debug("Health check passed", zap.String("timestamp", timestamp))
		w.WriteHeader(http.StatusOK) // Return 200 OK for healthy status
	}

	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GET /books?page=1&limit=10

// GetBooks godoc
// @Summary List all books
// @Description Get a paginated list of all books
// @Tags books
// @Accept       json
// @Produce      json
// @Param        requestBody    body      PaginationRequest    true   "Pagination and filter request"
// @Success      200      {object}  PaginationResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router /books/list [post]
func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) {
	var req PaginationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("binding failed", zap.Error(err))
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}
	books, pageCount, totalCount, err := h.repo.ListAllBooks(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to get books", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var booksResponse PaginationResponse
	booksResponse.TotalCount = totalCount
	booksResponse.PageCount = pageCount
	booksResponse.Data = books
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booksResponse)
}

// GET /books/{id}

// GetBookByID godoc
// @Summary Get book by ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} book.Book
// @Failure 404 {object} map[string]string
// @Router /books/{id} [get]
func (h *Handler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert ID to integer (if numeric IDs are used)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid book ID", zap.String("id", idStr))
		http.Error(w, "invalid book ID", http.StatusBadRequest)
		return
	}

	book, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
		} else {
			h.logger.Error("error retrieving book", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// POST /books

// CreateBook godoc
// @Summary Create a new book
// @Description Add a book to the library
// @Tags books
// @Accept json
// @Produce json
// @Param book body book.Book true "Book to create"
// @Success 201 {object} book.Book
// @Failure 400 {object} map[string]string
// @Router /books/create [post]
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.repo.Create(r.Context(), &b); err != nil {
		h.logger.Error("create failed", zap.Error(err))
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

// PUT /books/{id}

// UpdateBook godoc
// @Summary Update a book
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body book.Book true "Updated book"
// @Success 200 {object} book.Book
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /books/{id} [put]
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	b.ID = id

	if err := h.repo.Update(r.Context(), &b); err != nil {
		h.logger.Warn("update failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(b)
}

// DELETE /books/{id}

// DeleteBook godoc
// @Summary Delete a book
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Router /books/{id} [delete]
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.logger.Warn("delete failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
