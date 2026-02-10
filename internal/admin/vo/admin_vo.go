package vo

import (
	"time"

	"github.com/trv3wood/kuaizu-server/internal/models"
)

// AdminProjectVO is the admin-facing project response model.
type AdminProjectVO struct {
	ID                   int          `json:"id"`
	CreatorID            int          `json:"creatorId"`
	Name                 string       `json:"name"`
	Description          *string      `json:"description"`
	SchoolID             *int         `json:"schoolId"`
	Direction            *int         `json:"direction"`
	MemberCount          *int         `json:"memberCount"`
	Status               int          `json:"status"`
	PromotionStatus      int          `json:"promotionStatus"`
	PromotionExpireTime  *time.Time   `json:"promotionExpireTime"`
	ViewCount            int          `json:"viewCount"`
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
	SchoolName           *string      `json:"schoolName"`
	IsCrossSchool        *int         `json:"isCrossSchool"`
	EducationRequirement *int         `json:"educationRequirement"`
	SkillRequirement     *string      `json:"skillRequirement"`
	Creator              *AdminUserVO `json:"creator"`
}

// AdminUserVO is the admin-facing user response model.
type AdminUserVO struct {
	ID                  int        `json:"id"`
	OpenID              string     `json:"openId"`
	Nickname            *string    `json:"nickname"`
	Phone               *string    `json:"phone"`
	Email               *string    `json:"email"`
	SchoolID            *int       `json:"schoolId"`
	MajorID             *int       `json:"majorId"`
	Grade               *int       `json:"grade"`
	OliveBranchCount    int        `json:"oliveBranchCount"`
	FreeBranchUsedToday int        `json:"freeBranchUsedToday"`
	LastActiveDate      *time.Time `json:"lastActiveDate"`
	AuthStatus          int        `json:"authStatus"`
	AuthImgUrl          *string    `json:"authImgUrl"`
	EmailOptOut         bool       `json:"emailOptOut"`
	CreatedAt           time.Time  `json:"createdAt"`
	SchoolName          *string    `json:"schoolName"`
	SchoolCode          *string    `json:"schoolCode"`
	MajorName           *string    `json:"majorName"`
	ClassID             *int       `json:"classId"`
}

// AdminFeedbackVO is the admin-facing feedback response model.
type AdminFeedbackVO struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	Content      string    `json:"content"`
	ContactImage *string   `json:"contactImage"`
	Status       int       `json:"status"`
	AdminReply   *string   `json:"adminReply"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UserNickname *string   `json:"userNickname"`
}

// NewAdminProjectVO converts a Project model to AdminProjectVO.
func NewAdminProjectVO(p *models.Project) *AdminProjectVO {
	if p == nil {
		return nil
	}

	return &AdminProjectVO{
		ID:                   p.ID,
		CreatorID:            p.CreatorID,
		Name:                 p.Name,
		Description:          p.Description,
		SchoolID:             p.SchoolID,
		Direction:            p.Direction,
		MemberCount:          p.MemberCount,
		Status:               p.Status,
		PromotionStatus:      p.PromotionStatus,
		PromotionExpireTime:  p.PromotionExpireTime,
		ViewCount:            p.ViewCount,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		SchoolName:           p.SchoolName,
		IsCrossSchool:        p.IsCrossSchool,
		EducationRequirement: p.EducationRequirement,
		SkillRequirement:     p.SkillRequirement,
	}
}

// NewAdminUserVO converts a User model to AdminUserVO.
func NewAdminUserVO(u *models.User) *AdminUserVO {
	if u == nil {
		return nil
	}

	return &AdminUserVO{
		ID:                  u.ID,
		OpenID:              u.OpenID,
		Nickname:            u.Nickname,
		Phone:               u.Phone,
		Email:               u.Email,
		SchoolID:            u.SchoolID,
		MajorID:             u.MajorID,
		Grade:               u.Grade,
		OliveBranchCount:    u.OliveBranchCount,
		FreeBranchUsedToday: u.FreeBranchUsedToday,
		LastActiveDate:      u.LastActiveDate,
		AuthStatus:          u.AuthStatus,
		AuthImgUrl:          u.AuthImgUrl,
		EmailOptOut:         u.EmailOptOut,
		CreatedAt:           u.CreatedAt,
		SchoolName:          u.SchoolName,
		SchoolCode:          u.SchoolCode,
		MajorName:           u.MajorName,
		ClassID:             u.ClassID,
	}
}

// NewAdminFeedbackVO converts a Feedback model to AdminFeedbackVO.
func NewAdminFeedbackVO(f *models.Feedback) *AdminFeedbackVO {
	if f == nil {
		return nil
	}

	return &AdminFeedbackVO{
		ID:           f.ID,
		UserID:       f.UserID,
		Content:      f.Content,
		ContactImage: f.ContactImage,
		Status:       f.Status,
		AdminReply:   f.AdminReply,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		UserNickname: f.UserNickname,
	}
}
