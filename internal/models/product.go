package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// Product represents a product in the database
type Product struct {
	ID          int
	Name        string
	Type        int // 类型: 1-虚拟币, 2-服务权益
	Description *string
	Price       float64
	ConfigJson  *string // 配置参数(JSON格式)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ToVO converts Product to API ProductVO
func (p *Product) ToVO() *api.ProductVO {
	return &api.ProductVO{
		Id:          &p.ID,
		Name:        &p.Name,
		Type:        &p.Type,
		Description: p.Description,
		Price:       &p.Price,
		ConfigJson:  p.ConfigJson,
	}
}
