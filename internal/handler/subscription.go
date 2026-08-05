package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/programmerpark/subscription-aggregator/internal/model"
	"github.com/programmerpark/subscription-aggregator/internal/service"
)

type SubscriptionHandler struct {
	svc service.SubscriptionService
	log *slog.Logger
}

func NewSubscriptionHandler(svc service.SubscriptionService, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, log: log}
}

type errorResponse struct {
	Error string `json:"error"`
}

// Create godoc
// @Summary      Create subscription
// @Description  Creates a new subscription record
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      model.CreateSubscriptionRequest  true  "Subscription"
// @Success      201   {object}  model.Subscription
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if req.ServiceName == "" || req.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "service_name and user_id are required"})
		return
	}

	sub, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, sub)
}

// Get godoc
// @Summary      Get subscription by ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid subscription id"})
		return
	}

	sub, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, sub)
}

// List godoc
// @Summary      List subscriptions for a user
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  true   "User UUID (required)"
// @Param        service_name  query     string  false  "Filter by service name"
// @Param        limit         query     int     false  "Page size (default 20, max 100)"
// @Param        offset        query     int     false  "Offset (default 0)"
// @Success      200           {object}  model.ListResponse
// @Failure      400           {object}  errorResponse
// @Failure      500           {object}  errorResponse
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	userIDRaw := c.Query("user_id")
	if userIDRaw == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "user_id query parameter is required"})
		return
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid user_id"})
		return
	}

	filter := model.ListFilter{UserID: userID}
	if v := c.Query("service_name"); v != "" {
		filter.ServiceName = &v
	}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid limit"})
			return
		}
		filter.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid offset"})
			return
		}
		filter.Offset = n
	}

	result, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update godoc
// @Summary      Update subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "Subscription ID"
// @Param        body  body      model.UpdateSubscriptionRequest true  "Fields to update"
// @Success      200   {object}  model.Subscription
// @Failure      400   {object}  errorResponse
// @Failure      404   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid subscription id"})
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	sub, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, sub)
}

// Delete godoc
// @Summary      Delete subscription
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      204  "No Content"
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid subscription id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Sum godoc
// @Summary      Sum subscription costs for a period
// @Description  Calculates total cost as price * overlapping months for each matching subscription
// @Tags         subscriptions
// @Produce      json
// @Param        from          query     string  true   "Period start (MM-YYYY)"
// @Param        to            query     string  true   "Period end (MM-YYYY)"
// @Param        user_id       query     string  false  "Filter by user UUID"
// @Param        service_name  query     string  false  "Filter by service name"
// @Success      200           {object}  model.SumResponse
// @Failure      400           {object}  errorResponse
// @Failure      500           {object}  errorResponse
// @Router       /subscriptions/sum [get]
func (h *SubscriptionHandler) Sum(c *gin.Context) {
	fromRaw := c.Query("from")
	toRaw := c.Query("to")
	if fromRaw == "" || toRaw == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "from and to query params are required (MM-YYYY)"})
		return
	}

	from, err := model.ParseYearMonth(fromRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	to, err := model.ParseYearMonth(toRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	filter := model.SumFilter{From: from, To: to}
	if v := c.Query("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid user_id"})
			return
		}
		filter.UserID = &id
	}
	if v := c.Query("service_name"); v != "" {
		filter.ServiceName = &v
	}

	result, err := h.svc.Sum(c.Request.Context(), filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SubscriptionHandler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "subscription not found"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		h.log.Error("request failed", "error", err, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
