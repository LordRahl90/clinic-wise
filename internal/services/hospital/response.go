package hospital

import "clinic-wise/db/repositories"

type Response struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FromModel(m repositories.Hospital) *Response {
	return &Response{
		ID:   m.ID.String(),
		Name: m.Name.String,
	}
}
