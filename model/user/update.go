package user

import "time"

type UserUpdate struct {
	DisplayNameSet  bool
	DisplayName     *string
	PlanSet         bool
	Plan            *string
	BioSet          bool
	Bio             *string
	GenderSet       bool
	Gender          *string
	DateOfBirthSet  bool
	DateOfBirth     *time.Time
	ProfileImageSet bool
	ProfileImageURL *string
}
