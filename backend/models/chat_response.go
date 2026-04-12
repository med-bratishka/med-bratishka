package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChatResponse chat response
//
// swagger:model ChatResponse
type ChatResponse struct {
	ID          int64  `json:"id,omitempty"`
	DoctorID    int64  `json:"doctor_id,omitempty"`
	PatientID   int64  `json:"patient_id,omitempty"`
	OtherUserID int64  `json:"other_user_id,omitempty"`
	OtherLogin  string `json:"other_login,omitempty"`
	OtherName   string `json:"other_name,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

func (m *ChatResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *ChatResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *ChatResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *ChatResponse) UnmarshalBinary(b []byte) error {
	var res ChatResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
