package models

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// PaginationResponse pagination metadata
//
// swagger:model PaginationResponse
type PaginationResponse struct {
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`
	Total  int64 `json:"total,omitempty"`
}

func (m *PaginationResponse) Validate(formats strfmt.Registry) error { return nil }
func (m *PaginationResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
func (m *PaginationResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}
func (m *PaginationResponse) UnmarshalBinary(b []byte) error {
	var res PaginationResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
