package service

import (
	"context"
	"fmt"
	"log"

	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
	"github.com/trv3wood/kuaizu-server/internal/wechat"
)

// MessageService handles sending notifications (WeChat, etc.)
type MessageService struct {
	repo     *repository.Repository
	wxClient *wechat.Client
}

func NewMessageService(repo *repository.Repository) *MessageService {
	return &MessageService{
		repo:     repo,
		wxClient: wechat.NewClient(),
	}
}

// SendSubscribeMsgByBizKey sends a WeChat subscription message using a business key
func (s *MessageService) SendSubscribeMsgByBizKey(ctx context.Context, userID int, bizKey string, businessData map[string]string) error {
	// 1. Get user openid
	user, err := s.repo.User.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil || user.OpenID == "" {
		return fmt.Errorf("user not found or has no openid")
	}

	// 2. Get template config
	config, err := s.repo.MsgTemplate.GetByBizKey(ctx, bizKey)
	if err != nil {
		log.Printf("[MessageService.SendSubscribeMsgByBizKey] error getting config for %s: %v", bizKey, err)
		return fmt.Errorf("get template config: %w", err)
	}

	// 3. Send using client helper
	err = s.wxClient.SendByConfig(user.OpenID, config.TemplateID, config.ContentJSON, businessData)
	if err != nil {
		log.Printf("[MessageService.SendSubscribeMsgByBizKey] error sending message: %v", err)
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

// SyncSubscribeStatus syncs user subscription status from frontend
type TemplateSyncResult struct {
	BizKey string
	Result string // accept, reject, ban
}

func (s *MessageService) SyncSubscribeStatus(ctx context.Context, userID int, syncResults []TemplateSyncResult) error {
	for _, res := range syncResults {
		// 1. Get template_id by biz_key
		config, err := s.repo.MsgTemplate.GetByBizKey(ctx, res.BizKey)
		if err != nil {
			log.Printf("[MessageService.SyncSubscribeStatus] config not found for %s: %v", res.BizKey, err)
			continue
		}

		// 2. Map result to status
		var status models.SubscribeStatus
		switch res.Result {
		case "accept":
			status = models.SubscribeStatusAccept
		case "reject":
			status = models.SubscribeStatusReject
		default:
			status = models.SubscribeStatusReject // treat ban or other as reject
		}

		// 3. Upsert
		err = s.repo.SubscribeConfig.Upsert(ctx, &models.SubscribeConfig{
			UserID:     userID,
			BizKey:     res.BizKey,
			TemplateID: config.TemplateID,
			Status:     status,
		})
		if err != nil {
			log.Printf("[MessageService.SyncSubscribeStatus] upsert failed for %s: %v", res.BizKey, err)
		}
	}
	return nil
}
