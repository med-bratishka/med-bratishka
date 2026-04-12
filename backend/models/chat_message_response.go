package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChatMessageResponse chat message response
//
// swagger:model ChatMessageResponse
type ChatMessageResponse struct {
	ID                 int64  `json:"id,omitempty"`
	SenderID           int64  `json:"sender_id,omitempty"`
	SenderName         string `json:"sender_name,omitempty"`
	Content            string `json:"content,omitempty"`
	AttachmentURL      string `json:"attachment_url,omitempty"`
	AttachmentType     string `json:"attachment_type,omitempty"`
	AttachmentMimeType string `json:"attachment_mime_type,omitempty"`
	CreatedAt          int64  `json:"created_at,omitempty"`
}

func (m *ChatMessageResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *ChatMessageResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *ChatMessageResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *ChatMessageResponse) UnmarshalBinary(b []byte) error {
	var res ChatMessageResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
