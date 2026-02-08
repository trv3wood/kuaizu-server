package models

import "time"

// EmailPromotionStatus 邮件推广状态
type EmailPromotionStatus int

const (
	EmailPromotionStatusPending   EmailPromotionStatus = 0 // 待发送
	EmailPromotionStatusSending   EmailPromotionStatus = 1 // 发送中
	EmailPromotionStatusCompleted EmailPromotionStatus = 2 // 已完成
	EmailPromotionStatusFailed    EmailPromotionStatus = 3 // 失败
)

// EmailPromotion 邮件推广记录
type EmailPromotion struct {
	ID            int                  `db:"id"`
	OrderID       int                  `db:"order_id"`
	ProjectID     int                  `db:"project_id"`
	CreatorID     int                  `db:"creator_id"`
	MaxRecipients int                  `db:"max_recipients"` // 购买的最大发送人数
	TotalSent     int                  `db:"total_sent"`     // 实际发送数量
	Status        EmailPromotionStatus `db:"status"`         // 推广状态
	ErrorMessage  *string              `db:"error_message"`  // 错误信息
	StartedAt     *time.Time           `db:"started_at"`     // 开始发送时间
	CompletedAt   *time.Time           `db:"completed_at"`   // 完成时间
	CreatedAt     time.Time            `db:"created_at"`

	// Joined fields
	ProjectName *string  `db:"project_name"`
	Project     *Project `db:"-"`
}
