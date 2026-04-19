package schema

type SignUpUserBody struct {
	Username    string `json:"username" validate:"required"`
	Email       string `json:"email" validate:"required"`
	Password    string `json:"password" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
}
