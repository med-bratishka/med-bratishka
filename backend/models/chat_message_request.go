package models

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// ChatMessageRequest chat message request
//
// swagger:model ChatMessageRequest
type ChatMessageRequest struct {
	Content            *string `json:"content,omitempty"`
	AttachmentBase64   *string `json:"attachment_base64,omitempty"`
	AttachmentName     *string `json:"attachment_name,omitempty"`
	AttachmentMimeType *string `json:"attachment_mime_type,omitempty"`
	AttachmentType     *string `json:"attachment_type,omitempty"`
}

func (m *ChatMessageRequest) Validate(formats strfmt.Registry) error {
	var res []error
	if m.Content != nil {
		if err := validate.MinLength("content", "body", *m.Content, 1); err != nil {
			res = append(res, err)
		}
	}
	if m.AttachmentType != nil {
		if *m.AttachmentType != "image" && *m.AttachmentType != "audio" {
			res = append(res, errors.New(400, "attachment_type must be image or audio"))
		}
	}
	if (m.Content == nil || *m.Content == "") && (m.AttachmentBase64 == nil || *m.AttachmentBase64 == "") {
		res = append(res, errors.New(400, "either content or attachment_base64 is required"))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChatMessageRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

func (m *ChatMessageRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

func (m *ChatMessageRequest) UnmarshalBinary(b []byte) error {
	var res ChatMessageRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
