package exercise

type UpdateInput struct {
	NameSet          bool
	Name             *string
	CategorySet      bool
	Category         *string
	EquipmentSet     bool
	Equipment        *string
	MovementModeSet  bool
	MovementMode     *string
	IsMadeByAdminSet bool
	IsMadeByAdmin    *bool
	IsActiveSet      bool
	IsActive         *bool
}
