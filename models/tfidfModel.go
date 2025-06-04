package models

type TermFrequency struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	DocumentID uint    `gorm:"index" json:"document_id"`
	Word       string  `gorm:"size:100;index" json:"word"`
	Count      int     `json:"count"`     // сколько раз встретилось
	Frequency  float64 `json:"frequency"` // TF = term count / total terms

	Document Document `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

type InverseDocumentFrequency struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	CollectionID uint    `gorm:"index" json:"collection_id"`
	Word         string  `gorm:"size:100;index" json:"word"`
	IDFValue     float64 `json:"idf_value"` // логарифмическая мера

	Collection Collection `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}
