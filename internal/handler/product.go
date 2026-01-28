package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// ListProducts handles GET /products
func (s *Server) ListProducts(ctx echo.Context) error {
	products, err := s.repo.Product.GetAll(ctx.Request().Context())
	if err != nil {
		return InternalError(ctx, "获取商品列表失败")
	}

	// Convert to VOs
	var productVOs []api.ProductVO
	for _, product := range products {
		productVOs = append(productVOs, *product.ToVO())
	}

	return Success(ctx, productVOs)
}
