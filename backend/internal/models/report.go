package models

type DailySalesSummary struct {
	TotalSales     float64 `json:"total_sales"`
	NetSales       float64 `json:"net_sales"`
	VATAmount      float64 `json:"vat_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	SaleCount      int64   `json:"sale_count"`
}

type ProductSalesRow struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	SKU          string  `json:"sku"`
	QuantitySold float64 `json:"quantity_sold"`
	SalesAmount  float64 `json:"sales_amount"`
}

type PaymentMethodSummaryRow struct {
	PaymentMethod string  `json:"payment_method"`
	TotalAmount   float64 `json:"total_amount"`
}

type StockValuationRow struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    float64 `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
	StockValue  float64 `json:"stock_value"`
}

type DateRangeQuery struct {
	From string `form:"from" binding:"required,datetime=2006-01-02"`
	To   string `form:"to" binding:"required,datetime=2006-01-02"`
}
