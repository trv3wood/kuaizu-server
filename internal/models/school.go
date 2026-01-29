package models

import (
	"time"

	"github.com/trv3wood/kuaizu-server/api"
)

// School represents a school in the database
type School struct {
	ID         int
	SchoolName string
	SchoolCode *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ToVO converts School to API SchoolVO
func (s *School) ToVO() *api.SchoolVO {
	return &api.SchoolVO{
		Id:         &s.ID,
		SchoolName: &s.SchoolName,
		SchoolCode: s.SchoolCode,
	}
}
