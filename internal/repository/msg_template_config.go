package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

type MsgTemplateConfigRepository struct {
	db *sqlx.DB
}

func NewMsgTemplateConfigRepository(db *sqlx.DB) *MsgTemplateConfigRepository {
	return &MsgTemplateConfigRepository{db: db}
}

func (r *MsgTemplateConfigRepository) GetByBizKey(ctx context.Context, bizKey string) (*models.MsgTemplateConfig, error) {
	var config models.MsgTemplateConfig
	query := "SELECT biz_key, template_id, template_title, content_json, created_at, updated_at FROM msg_template_config WHERE biz_key = ?"
	err := r.db.GetContext(ctx, &config, query, bizKey)
	if err != nil {
		return nil, fmt.Errorf("get msg template config by biz_key: %w", err)
	}
	return &config, nil
}
