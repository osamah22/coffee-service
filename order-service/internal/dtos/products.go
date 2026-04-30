package dtos

type CreateProductRequest struct {
	Name         string `json:"name" binding:"required"`
	Category     string `json:"category" binding:"required,oneof=hot cold"`
	PriceInKurus int64  `json:"price_in_kurus" binding:"required,gte=0"`
	ImagePath    string `json:"image_path" binding:"required"`
	Available    *bool  `json:"available"`
}

type UpdateProductRequest struct {
	Name         string `json:"name" binding:"required"`
	Category     string `json:"category" binding:"required,oneof=hot cold"`
	PriceInKurus int64  `json:"price_in_kurus" binding:"required,gte=0"`
	ImagePath    string `json:"image_path" binding:"required"`
	Available    *bool  `json:"available"`
}
