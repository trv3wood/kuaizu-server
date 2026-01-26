package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// Project represents a project in the database
type Project struct {
	ID            int
	CreatorID     int
	Name          string
	Description   *string
	SchoolID      *int
	Direction     *int
	MemberCount   *int
	EducationReq  *int
	IsCrossSchool bool
	Status        int // 0-待审核, 1-已通过, 2-已驳回, 3-已关闭
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Joined fields
	SchoolName *string
	Creator    *User
}

// ProjectStatusToEnum converts integer status to API enum
func ProjectStatusToEnum(status int) api.ProjectStatus {
	switch status {
	case 0:
		return api.ProjectStatusPENDING
	case 1:
		return api.ProjectStatusACTIVE
	case 2:
		return api.ProjectStatusREJECTED
	case 3:
		return api.ProjectStatusCLOSED
	default:
		return api.ProjectStatusPENDING
	}
}

// ProjectStatusFromEnum converts API enum to integer status
func ProjectStatusFromEnum(status api.ProjectStatus) int {
	switch status {
	case api.ProjectStatusPENDING:
		return 0
	case api.ProjectStatusACTIVE:
		return 1
	case api.ProjectStatusREJECTED:
		return 2
	case api.ProjectStatusCLOSED:
		return 3
	default:
		return 0
	}
}

// ToVO converts Project to API ProjectVO
func (p *Project) ToVO() *api.ProjectVO {
	status := ProjectStatusToEnum(p.Status)

	return &api.ProjectVO{
		Id:            &p.ID,
		Name:          &p.Name,
		Description:   p.Description,
		SchoolId:      p.SchoolID,
		SchoolName:    p.SchoolName,
		Direction:     p.Direction,
		MemberCount:   p.MemberCount,
		EducationReq:  p.EducationReq,
		IsCrossSchool: &p.IsCrossSchool,
		Status:        &status,
		CreatedAt:     &p.CreatedAt,
	}
}

// ToDetailVO converts Project to API ProjectDetailVO
func (p *Project) ToDetailVO() *api.ProjectDetailVO {
	status := ProjectStatusToEnum(p.Status)

	vo := &api.ProjectDetailVO{
		Id:            &p.ID,
		Name:          &p.Name,
		Description:   p.Description,
		SchoolId:      p.SchoolID,
		SchoolName:    p.SchoolName,
		Direction:     p.Direction,
		MemberCount:   p.MemberCount,
		EducationReq:  p.EducationReq,
		IsCrossSchool: &p.IsCrossSchool,
		Status:        &status,
		CreatedAt:     &p.CreatedAt,
	}

	if p.Creator != nil {
		vo.Creator = p.Creator.ToVO()
	}

	return vo
}
