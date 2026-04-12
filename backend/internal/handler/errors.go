package handler

import (
	"errors"
	"net/http"

	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"
	"medbratishka/pkg/logs"
)

func makeErrorResponse(w http.ResponseWriter, r *http.Request, log logger.Logger, statusCode int, code, message string, cause error) {
	var serviceErr *service.ServiceError
	if errors.As(cause, &serviceErr) {
		statusCode = serviceErr.Code
		if serviceErr.PublicCode != "" {
			code = serviceErr.PublicCode
		}
		if serviceErr.PublicMsg != "" {
			message = serviceErr.PublicMsg
		}
	}

	errResp := &models.ErrorResponse{Code: code, Message: message, Type: getErrorType(statusCode)}

	rid := logs.RequestIDFromContext(r.Context())
	if rid != "" {
		errResp.Details = map[string]interface{}{"rid": rid}
	}

	if cause != nil {
		logHTTPError(log, r, statusCode, cause.Error())
	} else {
		logHTTPError(log, r, statusCode, message)
	}

	writeJSON(w, statusCode, errResp)
}
