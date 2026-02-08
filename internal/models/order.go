package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// Order represents an order in the database
type Order struct {
	ID         int
	UserID     int
	ActualPaid float64
	Status     int // 0-待支付, 1-已支付, 2-已取消, 3-已退款
	WxPayNo    *string
	PayTime    *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Joined/computed fields
	Items []*OrderItem
}

// OrderItem represents an order item in the database
type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Price     float64
	Quantity  int

	// Joined fields
	ProductName *string
}

// ToVO converts Order to API OrderVO
func (o *Order) ToVO() *api.OrderVO {
	status := api.OrderStatus(o.Status)

	// Get first item's product info for display
	var productID *int
	var productName *string
	if len(o.Items) > 0 {
		productID = &o.Items[0].ProductID
		productName = o.Items[0].ProductName
	}

	return &api.OrderVO{
		Id:          &o.ID,
		ProductId:   productID,
		ActualPaid:  &o.ActualPaid,
		Status:      &status,
		WxPayNo:     o.WxPayNo,
		PayTime:     o.PayTime,
		CreatedAt:   &o.CreatedAt,
		ProductName: productName,
	}
}
