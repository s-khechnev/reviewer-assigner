package users

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}
