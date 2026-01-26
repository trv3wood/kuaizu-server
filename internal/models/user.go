package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// User represents a user in the database
type User struct {
	ID               int
	OpenID           string
	Nickname         *string
	Phone            *string
	Email            *string
	SchoolID         *int
	MajorID          *int
	Grade            *int
	OliveBranchCount int
	StudentImgURL    *string
	AuthStatus       int // 0-未认证, 1-审核中, 2-已认证, 3-认证失败
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Joined fields (not always populated)
	SchoolName *string
	SchoolCode *string
	MajorName  *string
	ClassID    *int
}

// AuthStatusToEnum converts integer auth status to API enum
func AuthStatusToEnum(status int) api.AuthStatus {
	switch status {
	case 0:
		return api.AuthStatusUNVERIFIED
	case 1:
		return api.AuthStatusPENDING
	case 2:
		return api.AuthStatusVERIFIED
	case 3:
		return api.AuthStatusFAILED
	default:
		return api.AuthStatusUNVERIFIED
	}
}

// ToVO converts User to API UserVO
func (u *User) ToVO() *api.UserVO {
	authStatus := AuthStatusToEnum(u.AuthStatus)

	vo := &api.UserVO{
		Id:               &u.ID,
		Nickname:         u.Nickname,
		Phone:            u.Phone,
		Email:            u.Email,
		Grade:            u.Grade,
		OliveBranchCount: &u.OliveBranchCount,
		AuthStatus:       &authStatus,
		CreatedAt:        &u.CreatedAt,
	}

	// Populate school if available
	if u.SchoolID != nil && u.SchoolName != nil {
		vo.School = &api.SchoolVO{
			Id:         u.SchoolID,
			SchoolName: u.SchoolName,
			SchoolCode: u.SchoolCode,
		}
	}

	// Populate major if available
	if u.MajorID != nil && u.MajorName != nil {
		vo.Major = &api.MajorVO{
			Id:        u.MajorID,
			MajorName: u.MajorName,
			ClassId:   u.ClassID,
		}
	}

	return vo
}
