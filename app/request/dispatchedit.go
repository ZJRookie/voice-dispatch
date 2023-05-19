package request

type DispatchEditRequest struct {
	Id           int `json:"id"`
	ToPlatformId int `json:"to_platform_id" binding:"required"`
}
