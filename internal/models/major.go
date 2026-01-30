package models

import (
	"github.com/trv3wood/kuaizu-server/api"
)

// School represents a school in the database
type Major struct {
	Id        int
	ClassId   int
	MajorName string
}

// ToVO converts School to API SchoolVO
func (s *Major) ToVO() *api.MajorVO {
	return &api.MajorVO{
		Id:        &s.Id,
		ClassId:   &s.ClassId,
		MajorName: &s.MajorName,
	}
}

// MajorClass represents a major category in the database
type MajorClass struct {
	Id        int
	ClassName string
	Majors    []Major
}

// ToVO converts MajorClass to API MajorClassVO
func (mc *MajorClass) ToVO() *api.MajorClassVO {
	majors := make([]api.MajorVO, len(mc.Majors))
	for i, m := range mc.Majors {
		majors[i] = *m.ToVO()
	}

	return &api.MajorClassVO{
		Id:        &mc.Id,
		ClassName: &mc.ClassName,
		Majors:    &majors,
	}
}
