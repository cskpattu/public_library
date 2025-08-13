package book

type Book struct {
	ID     int    `json:"id" example:"1"`
	Title  string `json:"title" example:"The Great Gatsby"`
	Author string `json:"author" example:"F. Scott Fitzgerald"`
	ISBN   string `json:"isbn" example:"9780743273565"`
}

// PaginationRequest represents a request for paginated data with search
type PaginationRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search"`
}

// Sort represents sorting options for queries
type Sort struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

type BookResponse struct {
	ID     int    `json:"id" example:"1"`
	Title  string `json:"title" example:"The Great Gatsby"`
	Author string `json:"author" example:"F. Scott Fitzgerald"`
	ISBN   string `json:"isbn" example:"9780743273565"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	TotalCount int64       `json:"total_count"`
	PageCount  int64       `json:"page_count"`
	Data       interface{} `json:"data"`
}

// StatusResponse represents the health check response
type StatusResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}
