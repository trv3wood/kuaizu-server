package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// Product represents a product in the database
type Product struct {
	ID              int
	Name            string
	Description     *string
	Price           float64
	AvailableAmount int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ToVO converts Product to API ProductVO
func (p *Product) ToVO() *api.ProductVO {
	return &api.ProductVO{
		Id:              &p.ID,
		Name:            &p.Name,
		Description:     p.Description,
		Price:           &p.Price,
		AvailableAmount: &p.AvailableAmount,
	}
}
