package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// Order represents an order in the database
type Order struct {
	ID         int        `db:"id"`
	UserID     int        `db:"user_id"`
	ActualPaid float64    `db:"actual_paid"`
	Status     int        `db:"status"` // 0-待支付, 1-已支付, 2-已取消, 3-已退款
	WxPayNo    *string    `db:"wx_pay_no"`
	PayTime    *time.Time `db:"pay_time"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`

	// Joined/computed fields
	Items []*OrderItem `db:"-"`
}

// OrderItem represents an order item in the database
type OrderItem struct {
	ID        int     `db:"id"`
	OrderID   int     `db:"order_id"`
	ProductID int     `db:"product_id"`
	Price     float64 `db:"price"`
	Quantity  int     `db:"quantity"`

	// Joined fields
	ProductName *string `db:"product_name"`
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
