package dto

type CreateCollectionReq struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCollectionReq struct {
	Name string `json:"name" binding:"required"`
}

type AddDocumentToCollectionsReq struct {
	DocumentID    uint   `json:"document_id" binding:"required"`
	CollectionIDs []uint `json:"collection_ids" binding:"required"`
}
