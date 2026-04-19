package handler

import (
	"fmt"
	"net/http"

	"medbratishka/pkg/logger"
	"medbratishka/pkg/logs"
)

func logHTTPError(log logger.Logger, r *http.Request, status int, errText string) {
	if log == nil || r == nil {
		return
	}

	rid := logs.RequestIDFromContext(r.Context())
	msg := fmt.Sprintf("Code: %d Method: %s URL: %s error: %s rid: %s", status, r.Method, r.URL.String(), errText, rid)
	if status >= http.StatusInternalServerError {
		log.Errorf(msg)
		return
	}
	log.Warningf(msg)
}
