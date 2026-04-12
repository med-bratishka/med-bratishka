package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// PaginatedChatMessagesResponse chat messages with pagination
//
// swagger:model PaginatedChatMessagesResponse
type PaginatedChatMessagesResponse struct {
	Items      []*ChatMessageResponse `json:"items"`
	Pagination *PaginationResponse    `json:"pagination"`
}

func (m *PaginatedChatMessagesResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *PaginatedChatMessagesResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *PaginatedChatMessagesResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *PaginatedChatMessagesResponse) UnmarshalBinary(b []byte) error {
	var res PaginatedChatMessagesResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
