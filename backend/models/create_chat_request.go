package models

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// CreateChatRequest create chat request
//
// swagger:model CreateChatRequest
type CreateChatRequest struct {
	// doctor id
	// Required: true
	DoctorID *int64 `json:"doctor_id"`
}

func (m *CreateChatRequest) Validate(formats strfmt.Registry) error {
	if err := validate.Required("doctor_id", "body", m.DoctorID); err != nil {
		return errors.CompositeValidationError(err)
	}
	return nil
}

func (m *CreateChatRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

func (m *CreateChatRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

func (m *CreateChatRequest) UnmarshalBinary(b []byte) error {
	var res CreateChatRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
