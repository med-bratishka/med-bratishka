package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// PaginatedChatsResponse chats response with pagination
//
// swagger:model PaginatedChatsResponse
type PaginatedChatsResponse struct {
	Items      []*ChatResponse     `json:"items"`
	Pagination *PaginationResponse `json:"pagination"`
}

func (m *PaginatedChatsResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *PaginatedChatsResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *PaginatedChatsResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *PaginatedChatsResponse) UnmarshalBinary(b []byte) error {
	var res PaginatedChatsResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
