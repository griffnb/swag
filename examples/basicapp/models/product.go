package models

// Product represents a product in the system
type Product struct {
	ID          int     `json:"id" example:"1"`
	Name        string  `json:"name" example:"Widget"`
	Price       float64 `json:"price" example:"19.99"`
	Description string  `json:"description" example:"A useful widget"`
}
