package models

import "time"

type Category struct {
	ID          string     `json:"id"`
	BranchID    string     `json:"branch_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
	UpdatedBy   *string    `json:"updated_by,omitempty"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=150"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

type UpdateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=150"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

type Product struct {
	ID             string     `json:"id"`
	BranchID       string     `json:"branch_id"`
	CategoryID     string     `json:"category_id"`
	Name           string     `json:"name"`
	SKU            string     `json:"sku"`
	Barcode        *string    `json:"barcode,omitempty"`
	CostPrice      float64    `json:"cost_price"`
	SellingPrice   float64    `json:"selling_price"`
	VATPercentage  float64    `json:"vat_percentage"`
	IsVATInclusive bool       `json:"is_vat_inclusive"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	CreatedBy      *string    `json:"created_by,omitempty"`
	UpdatedBy      *string    `json:"updated_by,omitempty"`
}

type CreateProductRequest struct {
	CategoryID     string   `json:"category_id" binding:"required,uuid"`
	Name           string   `json:"name" binding:"required,min=2,max=200"`
	SKU            string   `json:"sku" binding:"required,min=1,max=100"`
	Barcode        *string  `json:"barcode" binding:"omitempty,max=100"`
	CostPrice      float64  `json:"cost_price" binding:"required,gte=0"`
	SellingPrice   float64  `json:"selling_price" binding:"required,gte=0"`
	VATPercentage  *float64 `json:"vat_percentage" binding:"omitempty,gte=0"`
	IsVATInclusive bool     `json:"is_vat_inclusive"`
	IsActive       *bool    `json:"is_active"`
}

type UpdateProductRequest struct {
	CategoryID     string   `json:"category_id" binding:"required,uuid"`
	Name           string   `json:"name" binding:"required,min=2,max=200"`
	SKU            string   `json:"sku" binding:"required,min=1,max=100"`
	Barcode        *string  `json:"barcode" binding:"omitempty,max=100"`
	CostPrice      float64  `json:"cost_price" binding:"required,gte=0"`
	SellingPrice   float64  `json:"selling_price" binding:"required,gte=0"`
	VATPercentage  *float64 `json:"vat_percentage" binding:"omitempty,gte=0"`
	IsVATInclusive bool     `json:"is_vat_inclusive"`
	IsActive       bool     `json:"is_active"`
}

type Supplier struct {
	ID           string     `json:"id"`
	BranchID     string     `json:"branch_id"`
	Name         string     `json:"name"`
	MobileNumber *string    `json:"mobile_number,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Address      *string    `json:"address,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateSupplierRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=150"`
	MobileNumber *string `json:"mobile_number" binding:"omitempty,max=20"`
	Email        *string `json:"email" binding:"omitempty,email,max=255"`
	Address      *string `json:"address" binding:"omitempty,max=1000"`
}

type UpdateSupplierRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=150"`
	MobileNumber *string `json:"mobile_number" binding:"omitempty,max=20"`
	Email        *string `json:"email" binding:"omitempty,email,max=255"`
	Address      *string `json:"address" binding:"omitempty,max=1000"`
}

type Customer struct {
	ID            string     `json:"id"`
	BranchID      string     `json:"branch_id"`
	Name          string     `json:"name"`
	MobileNumber  *string    `json:"mobile_number,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Address       *string    `json:"address,omitempty"`
	LoyaltyPoints float64    `json:"loyalty_points"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type CreateCustomerRequest struct {
	Name          string   `json:"name" binding:"required,min=2,max=150"`
	MobileNumber  *string  `json:"mobile_number" binding:"omitempty,max=20"`
	Email         *string  `json:"email" binding:"omitempty,email,max=255"`
	Address       *string  `json:"address" binding:"omitempty,max=1000"`
	LoyaltyPoints *float64 `json:"loyalty_points" binding:"omitempty,gte=0"`
}

type UpdateCustomerRequest struct {
	Name          string  `json:"name" binding:"required,min=2,max=150"`
	MobileNumber  *string `json:"mobile_number" binding:"omitempty,max=20"`
	Email         *string `json:"email" binding:"omitempty,email,max=255"`
	Address       *string `json:"address" binding:"omitempty,max=1000"`
	LoyaltyPoints float64 `json:"loyalty_points" binding:"required,gte=0"`
}

type Sale struct {
	ID               string            `json:"id"`
	BranchID         string            `json:"branch_id"`
	UserID           string            `json:"user_id"`
	CustomerID       *string           `json:"customer_id,omitempty"`
	TotalAmount      float64           `json:"total_amount"`
	VATAmount        float64           `json:"vat_amount"`
	DiscountAmount   float64           `json:"discount_amount"`
	NetAmount        float64           `json:"net_amount"`
	CreatedAt        time.Time         `json:"created_at"`
	IdempotencyKey   *string           `json:"idempotency_key,omitempty"`
	Items            []SaleItem        `json:"items"`
	AppliedDiscounts []AppliedDiscount `json:"applied_discounts,omitempty"`
	Payments         []SalePayment     `json:"payments"`
}

type SaleItem struct {
	SaleID    string  `json:"sale_id"`
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	VATAmount float64 `json:"vat_amount"`
	LineTotal float64 `json:"line_total"`
}

type SalePayment struct {
	ID            string    `json:"id"`
	SaleID        string    `json:"sale_id"`
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	PaidAt        time.Time `json:"paid_at"`
}

type CreateSaleItemRequest struct {
	ProductID          string  `json:"product_id" binding:"required,uuid"`
	Quantity           float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice          float64 `json:"unit_price" binding:"required,gte=0"`
	VATAmount          float64 `json:"vat_amount" binding:"required,gte=0"`
	LineTotal          float64 `json:"line_total" binding:"required,gte=0"`
	ItemDiscountAmount float64 `json:"item_discount_amount" binding:"omitempty,gte=0"`
}

type CreateSalePaymentRequest struct {
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=CASH CARD QR"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}

type CreateSaleRequest struct {
	CustomerID          *string                    `json:"customer_id" binding:"omitempty,uuid"`
	TotalAmount         float64                    `json:"total_amount" binding:"required,gte=0"`
	VATAmount           float64                    `json:"vat_amount" binding:"required,gte=0"`
	DiscountAmount      float64                    `json:"discount_amount" binding:"required,gte=0"`
	NetAmount           float64                    `json:"net_amount" binding:"required,gte=0"`
	Items               []CreateSaleItemRequest    `json:"items" binding:"required,min=1,dive"`
	Payments            []CreateSalePaymentRequest `json:"payments" binding:"required,min=1,dive"`
	DiscountRuleIDs     []string                   `json:"discount_rule_ids" binding:"omitempty,dive,uuid"`
	LoyaltyRedeemPoints float64                    `json:"loyalty_redeem_points" binding:"omitempty,gte=0"`
	IdempotencyKey      *string                    `json:"idempotency_key" binding:"omitempty,max=100"`
}

type AppliedDiscount struct {
	DiscountRuleID *string `json:"discount_rule_id,omitempty"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Amount         float64 `json:"amount"`
}
