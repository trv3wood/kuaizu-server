package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/service"
)

// SyncUserSubscription handles POST /user/subscribe
func (s *Server) SyncUserSubscription(ctx echo.Context) error {
	var req api.SubscribeSyncRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.BaseResponse{
			Code:    http.StatusBadRequest,
			Message: "参数解析失败",
		})
	}

	userID := GetUserID(ctx)
	syncResults := make([]service.TemplateSyncResult, len(req.Templates))
	for i, t := range req.Templates {
		syncResults[i] = service.TemplateSyncResult{
			BizKey: t.BizKey,
			Result: string(t.Result),
		}
	}

	err := s.svc.Message.SyncSubscribeStatus(ctx.Request().Context(), userID, syncResults)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "同步订阅状态失败",
		})
	}

	return ctx.JSON(http.StatusOK, api.BaseResponse{
		Code:    http.StatusOK,
		Message: "同步成功",
	})
}
