package dtos

type UpdateBookDto struct {
	Title string `json:"title,omitempty" bson:"title,omitempty"`
	Author string `json:"author,omitempty" bson:"author,omitempty"`
	Year string `json:"year,omitempty" bson:"year,omitempty"`
}