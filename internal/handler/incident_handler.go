package handler

import (
	"encoding/json"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/service"

	"github.com/theartofdevel/logging"
)

// @Summary      Создание инцидента
// @Description  Создает новую опасную зону с географическими координатами
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        incident  body      IncidentJSON  true  "Данные инцидента"
// @Success      201       {object}  incedentRequestResponse
// @Failure      400       {object}  badRequestErrorResponse
// @Failure      401       {object}  unauthorizedErrorResponse
// @Failure      500       {object}  internalServerErrorResponse
// @Security     ApiKeyAuth
// @Router       /incidents [post]
func (h *Handler) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var req IncidentJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := service.CreateIncidentRequestInput{
		Title:       req.Title,
		Description: req.Description,
		Lat:         req.Lat,
		Long:        req.Long,
		Radius:      req.Radius,
		Active:      req.Active,
	}

	out, err := h.svc.CreateIncident(r.Context(), &in)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary      Получение списка инцидентов
// @Description  Получает список инцидентов с пагинацией
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Номер страницы"  default(1)
// @Param        limit  query     int  false  "Количество на странице"  default(5)
// @Success      200    {object}  service.PaginateIncidentsOutput
// @Failure      400    {object}  badRequestErrorResponsePaginate
// @Failure      401    {object}  unauthorizedErrorResponse
// @Security     ApiKeyAuth
// @Router       /incidents [get]
func (h *Handler) handlePaginate(w http.ResponseWriter, r *http.Request) {
	rawPage := r.URL.Query().Get("page")
	rawlimit := r.URL.Query().Get("limit")

	out, err := h.svc.PaginateIncident(r.Context(), rawlimit, rawPage)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// @Summary      Получение инцидента по ID
// @Description  Получает информацию об инциденте по его идентификатору
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID инцидента"
// @Success      200  {object}  incedentRequestResponse
// @Failure      400  {object}  badRequestErrorResponseGetByID
// @Failure      401  {object}  unauthorizedErrorResponse
// @Failure      404  {object}  notFoundErrorResponse
// @Security     ApiKeyAuth
// @Router       /incidents/{id} [get]
func (h *Handler) handleGetIncidentByID(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	out, err := h.svc.GetIncidentByID(r.Context(), rawID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Обновление инцидента
// @Description  Полностью обновляет данные инцидента
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        id        path      int           true  "ID инцидента"
// @Param        incident  body      IncidentJSON  true  "Обновленные данные инцидента"
// @Success      200       {object}  incedentRequestResponse
// @Failure      400       {object}  badRequestErrorResponse
// @Failure      401       {object}  unauthorizedErrorResponse
// @Failure      404       {object}  notFoundErrorResponse
// @Security     ApiKeyAuth
// @Router       /incidents/{id} [put]
func (h *Handler) handlePutIncident(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	var req IncidentJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := &service.FullUpdateIncidentRequestInput{
		ID:          rawID,
		Title:       req.Title,
		Description: req.Description,
		Lat:         req.Lat,
		Long:        req.Long,
		Radius:      req.Radius,
		Active:      req.Active,
	}

	out, err := h.svc.FullUpdateIncident(r.Context(), in)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Удаление инцидента
// @Description  Удаляет инцидент по его идентификатору
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID инцидента"
// @Success      200
// @Failure      400  {object}  badRequestErrorResponseGetByID
// @Failure      401  {object}  unauthorizedErrorResponse
// @Failure      404  {object}  notFoundErrorResponse
// @Security     ApiKeyAuth
// @Router       /incidents/{id} [delete]
func (h *Handler) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	if err := h.svc.DeleteIncident(r.Context(), rawID); err != nil {
		h.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
