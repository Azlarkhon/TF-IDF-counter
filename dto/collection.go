package dto

type CreateCollectionReq struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCollectionReq struct {
	Name string `json:"name" binding:"required"`
}
