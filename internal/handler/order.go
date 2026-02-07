package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
	"github.com/trv3wood/kuaizu-server/internal/wechat"
)

// ListMyOrders handles GET /orders/me
func (s *Server) ListMyOrders(ctx echo.Context, params api.ListMyOrdersParams) error {
	userID := GetUserID(ctx)

	// Build list params
	listParams := repository.OrderListParams{
		UserID: userID,
		Page:   1,
		Size:   10,
	}

	if params.Page != nil {
		listParams.Page = int(*params.Page)
	}
	if params.Size != nil {
		listParams.Size = int(*params.Size)
	}
	if listParams.Page < 1 {
		listParams.Page = 1
	}
	if listParams.Size < 1 || listParams.Size > 100 {
		listParams.Size = 10
	}

	if params.Status != nil {
		status := int(*params.Status)
		listParams.Status = &status
	}

	// Query
	orders, total, err := s.repo.Order.ListByUserID(ctx.Request().Context(), listParams)
	if err != nil {
		return InternalError(ctx, "获取订单列表失败")
	}

	// Convert to VOs
	list := make([]api.OrderVO, len(orders))
	for i, o := range orders {
		list[i] = *o.ToVO()
	}

	// Build pagination info
	totalPages := int((total + int64(listParams.Size) - 1) / int64(listParams.Size))
	pageInfo := api.PageInfo{
		Page:       &listParams.Page,
		Size:       &listParams.Size,
		Total:      &total,
		TotalPages: &totalPages,
	}

	return Success(ctx, struct {
		List     *[]api.OrderVO `json:"list"`
		PageInfo *api.PageInfo  `json:"pageInfo"`
	}{
		List:     &list,
		PageInfo: &pageInfo,
	})
}

// CreateOrder handles POST /orders
func (s *Server) CreateOrder(ctx echo.Context) error {
	userID := GetUserID(ctx)

	var reqItems []api.CreateOrderDTO
	if err := ctx.Bind(&reqItems); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if len(reqItems) == 0 {
		return BadRequest(ctx, "订单商品不能为空")
	}

	// Build order items and calculate total
	var actualPaid float64
	var orderItems []*models.OrderItem

	for _, item := range reqItems {
		if item.ProductId <= 0 {
			return BadRequest(ctx, "商品ID无效")
		}
		if item.Quantity <= 0 {
			return BadRequest(ctx, "购买数量必须大于0")
		}

		// Get product
		product, err := s.repo.Product.GetByID(ctx.Request().Context(), item.ProductId)
		if err != nil {
			return InternalError(ctx, "获取商品信息失败")
		}
		if product == nil {
			return NotFound(ctx, fmt.Sprintf("商品ID %d 不存在", item.ProductId))
		}

		// Add to total
		actualPaid += product.Price * float64(item.Quantity)

		// Create order item
		orderItems = append(orderItems, &models.OrderItem{
			ProductID:   item.ProductId,
			Price:       product.Price,
			Quantity:    item.Quantity,
			ProductName: &product.Name,
		})
	}

	// Create order with items
	order := &models.Order{
		UserID:     userID,
		ActualPaid: actualPaid,
		Status:     0, // 待支付
		Items:      orderItems,
	}

	createdOrder, err := s.repo.Order.Create(ctx.Request().Context(), order)
	if err != nil {
		return InternalError(ctx, "创建订单失败")
	}

	return Success(ctx, createdOrder.ToVO())
}

// GetOrder handles GET /orders/{id}
func (s *Server) GetOrder(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)

	order, err := s.repo.Order.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取订单详情失败")
	}
	if order == nil {
		return NotFound(ctx, "订单不存在")
	}

	// Verify ownership
	if order.UserID != userID {
		return Forbidden(ctx, "无权查看此订单")
	}

	return Success(ctx, order.ToVO())
}

// InitiatePayment handles POST /orders/{id}/pay
func (s *Server) InitiatePayment(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)
	openID := GetOpenID(ctx)

	if openID == "" {
		return BadRequest(ctx, "无法获取用户OpenID")
	}

	order, err := s.repo.Order.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取订单详情失败")
	}
	if order == nil {
		return NotFound(ctx, "订单不存在")
	}

	// Verify ownership
	if order.UserID != userID {
		return Forbidden(ctx, "无权操作此订单")
	}

	// Check order status
	if order.Status != 0 {
		return BadRequest(ctx, "订单状态不允许支付")
	}

	// Get WeChat Pay config
	payConfig, err := wechat.DefaultPayConfig()
	if err != nil {
		return InternalError(ctx, "支付配置错误: "+err.Error())
	}

	payClient, err := wechat.NewPayClient(payConfig)
	if err != nil {
		return InternalError(ctx, "初始化支付客户端失败: "+err.Error())
	}

	// Build description from order items
	description := "快组商品购买"
	if len(order.Items) > 0 && order.Items[0].ProductName != nil {
		description = *order.Items[0].ProductName
	}

	// Use order ID as out_trade_no
	outTradeNo := wechat.GenerateOutTradeNo(order.ID)

	// Amount in cents
	amountCents := int(order.ActualPaid * 100)

	// Create prepay order and get payment params directly
	paymentParams, err := payClient.CreatePrepayOrderWithPayment(
		ctx.Request().Context(),
		outTradeNo,
		description,
		openID,
		amountCents,
	)
	if err != nil {
		return InternalError(ctx, "创建支付订单失败: "+err.Error())
	}

	return Success(ctx, api.WechatPaymentParams{
		TimeStamp: &paymentParams.TimeStamp,
		NonceStr:  &paymentParams.NonceStr,
		Package:   &paymentParams.Package,
		SignType:  &paymentParams.SignType,
		PaySign:   &paymentParams.PaySign,
	})
}
