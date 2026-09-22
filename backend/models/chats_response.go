package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChatsResponse chats response
//
// swagger:model ChatsResponse
type ChatsResponse struct {
	Chats []*ChatResponse `json:"chats"`
}

func (m *ChatsResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *ChatsResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *ChatsResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *ChatsResponse) UnmarshalBinary(b []byte) error {
	var res ChatsResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
