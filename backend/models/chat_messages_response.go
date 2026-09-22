package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChatMessagesResponse chat messages response
//
// swagger:model ChatMessagesResponse
type ChatMessagesResponse struct {
	Messages []*ChatMessageResponse `json:"messages"`
}

func (m *ChatMessagesResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *ChatMessagesResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *ChatMessagesResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *ChatMessagesResponse) UnmarshalBinary(b []byte) error {
	var res ChatMessagesResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
