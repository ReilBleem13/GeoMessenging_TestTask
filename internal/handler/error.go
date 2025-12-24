package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeAPIResponse(w http.ResponseWriter, status int, code, message string) {
	resp := apiErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
}

func (h *Handler) WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		h.logger.Error("application error",
			logging.StringAttr("code", string(appErr.Code)),
			logging.StringAttr("message", appErr.Message),
		)
		writeAPIResponse(w, statusFromCode(appErr.Code), string(appErr.Code), appErr.Message)
		return
	}
	h.logger.Error("unexpected error", logging.ErrAttr(err))
	writeAPIResponse(w, http.StatusInternalServerError, "SERVER_ERROR", "internal server error")
}

func statusFromCode(code domain.ErrorCode) int {
	switch code {
	case domain.CodeIncedentExists:
		return http.StatusBadRequest
	default:
		return 503
		// подумать
	}
}
