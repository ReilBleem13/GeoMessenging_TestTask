package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

// apiErrorResponse представляет структуру ответа об ошибке API
// @Description Структура ошибки API
type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if resp != nil {
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func writeAPIResponse(w http.ResponseWriter, status int, code, message string) {
	resp := apiErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	writeJSON(w, status, resp)
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
	case domain.CodeInvalidRequest, domain.CodeInvalidValidation:
		return 400
	case domain.CodeAlreadyExists:
		return 409
	case domain.CodeNotFound:
		return 404
	default:
		return 503
	}
}

// badRequestErrorResponse представляет структуру ответа об ошибке 400
// @Description Ошибка валидации или некорректного запроса
type badRequestErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"INVALID_VALIDATION"`
		Message string `json:"message" example:"lat or long is invalid"`
	} `json:"error"`
}

// unauthorizedErrorResponse представляет структуру ответа об ошибке 401
// @Description Ошибка аутентификации
type unauthorizedErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"UNAUTHORIZED"`
		Message string `json:"message" example:"invalid api key"`
	} `json:"error"`
}

// internalServerErrorResponse представляет структуру ответа об ошибке 500
// @Description Внутренняя ошибка сервера
type internalServerErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"SERVER_ERROR"`
		Message string `json:"message" example:"internal server error"`
	} `json:"error"`
}

// badRequestErrorResponsePaginate представляет структуру ответа об ошибке 400 для пагинации
// @Description Ошибка валидации или некорректного запроса
type badRequestErrorResponsePaginate struct {
	Error struct {
		Code    string `json:"code" example:"INVALID_VALIDATION"`
		Message string `json:"message" example:"invalid limit format, must be integer"`
	} `json:"error"`
}

// badRequestErrorResponseGetByID представляет структуру ответа об ошибке 400 для получения по айди
// @Description Ошибка валидации или некорректного запроса
type badRequestErrorResponseGetByID struct {
	Error struct {
		Code    string `json:"code" example:"INVALID_VALIDATION"`
		Message string `json:"message" example:"invalid id format, must be integer"`
	} `json:"error"`
}

// notFoundErrorResponse представляет структуру ответа об ошибке 404
// @Description Ошибка получения, сущность не найдена
type notFoundErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"NOT_FOUND"`
		Message string `json:"message" example:"incident is not exists"`
	} `json:"error"`
}
