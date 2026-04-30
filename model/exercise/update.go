package exercise

type UpdateInput struct {
	NameSet             bool
	Name                *string
	SlugSet             bool
	Slug                *string
	CategorySet         bool
	Category            *string
	EquipmentSet        bool
	Equipment           *string
	PrimaryMuscleSet    bool
	PrimaryMuscle       *string
	SecondaryMusclesSet bool
	SecondaryMuscles    []string
	DifficultySet       bool
	Difficulty          *string
	MovementModeSet     bool
	MovementMode        *string
	IsMadeByAdminSet    bool
	IsMadeByAdmin       *bool
	IsActiveSet         bool
	IsActive            *bool
}
