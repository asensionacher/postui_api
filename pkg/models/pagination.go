package models

// PaginatedProductResponse represents a paginated response
type PaginatedProductResponse struct {
	Data       []Product  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains pagination information
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
