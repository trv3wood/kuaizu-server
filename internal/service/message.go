package service

import (
	"context"
	"fmt"
	"log"

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
