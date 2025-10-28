package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"tz/internal/domain"
	"tz/internal/dto"
	"tz/pkg/valid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type SubscriptionServiceI interface {
	CreateSubscription(ctx context.Context, sub dto.CreateSubscriptionRequest) (dto.SubscriptionOutput, error)
	SubscriptionByID(ctx context.Context, id uuid.UUID) (dto.SubscriptionOutput, error)
	Subscriptions(ctx context.Context, filter dto.SubscriptionFilter) (dto.SubscriptionsOutput, error)
	SubscriptionsCost(ctx context.Context, req dto.CostRequest) (int, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, sub dto.UpdateSubscriptionRequest) (dto.SubscriptionOutput, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type SubscriptionHandler struct {
	service SubscriptionServiceI
	log     *zap.Logger
}

func NewHandler(service SubscriptionServiceI, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{service: service, log: log}
}

func (h *SubscriptionHandler) Init() *gin.Engine {
	router := gin.New()

	router.Use(
		gin.Recovery(),
		h.logging(),
	)

	h.initAPI(router)

	return router

}

func (h *SubscriptionHandler) initAPI(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	subscriptions := router.Group("/subscriptions")
	{
		subscriptions.POST("/", h.createSubscription)
		subscriptions.GET("/", h.listSubscriptions)
		subscriptions.GET("/cost", h.subscriptionsCost)
		subscriptions.GET("/:id", h.subscription)
		subscriptions.PATCH("/:id", h.updateSubscription)
		subscriptions.DELETE("/:id", h.deleteSubscription)
	}
}

// @Summary Create a new subscription
// @Description Create a subscription record for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body dto.CreateSubscriptionRequest true "Subscription data"
// @Success 200 {object} dto.SubscriptionOutput
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) createSubscription(c *gin.Context) {
	log := h.loggerWith(c)
	var req dto.CreateSubscriptionRequest
	if err := c.BindJSON(&req); err != nil {
		log.Warn("Failed to bind create subscription request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := valid.ValidateStruct(req); err != nil {
		log.Warn("Validation failed for create subscription", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.CreateSubscription(c.Request.Context(), req)
	if err != nil {
		log.Error("Failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// @Summary Get list of subscriptions
// @Description Get paginated list of subscriptions with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10) maximum(100)
// @Success 200 {object} dto.SubscriptionsOutput
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) listSubscriptions(c *gin.Context) {
	log := h.loggerWith(c)
	var filter dto.SubscriptionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Warn("Failed to bind subscription filter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	subscriptions, err := h.service.Subscriptions(c.Request.Context(), filter)
	if err != nil {
		log.Error("Failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// @Summary Calculate total subscription cost
// @Description Calculate total cost of subscriptions for a given period and filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Param from query string false "From in MM-YYYY format"
// @Param to query string false "To in MM-YYYY format"
// @Success 200 {integer} int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/cost [get]
func (h *SubscriptionHandler) subscriptionsCost(c *gin.Context) {
	log := h.loggerWith(c)
	var request dto.CostRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Warn("Failed to bind cost request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	subscriptionsCost, err := h.service.SubscriptionsCost(c.Request.Context(), request)
	if err != nil {
		log.Error("Failed to calculate total cost", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subscriptionsCost)
}

// @Summary Get subscription by ID
// @Description Retrieve a single subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} dto.SubscriptionOutput
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) subscription(c *gin.Context) {
	log := h.loggerWith(c)
	id, err := h.parseID(c, "id")
	if err != nil {
		log.Warn("Invalid subscription ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	subscription, err := h.service.SubscriptionByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Subscription not found", zap.String("id", id.String()))
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		log.Error("Failed to get subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// @Summary Update a subscription
// @Description Partially update a subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body dto.UpdateSubscriptionRequest true "Fields to update"
// @Success 200 {object} dto.SubscriptionOutput
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [patch]
func (h *SubscriptionHandler) updateSubscription(c *gin.Context) {
	log := h.loggerWith(c)
	id, err := h.parseID(c, "id")
	if err != nil {
		log.Warn("Invalid subscription ID in update", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}
	var req dto.UpdateSubscriptionRequest
	if err := c.BindJSON(&req); err != nil {
		log.Warn("Failed to bind update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := valid.ValidateStruct(req); err != nil {
		log.Warn("Validation failed for update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.UpdateSubscription(c.Request.Context(), id, req)
	if err != nil {
		log.Error("Failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// @Summary Delete a subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) deleteSubscription(c *gin.Context) {
	log := h.loggerWith(c)

	id, err := h.parseID(c, "id")
	if err != nil {
		log.Warn("Invalid subscription ID in delete", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.service.DeleteSubscription(c.Request.Context(), id); err != nil {
		log.Error("Failed to delete subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SubscriptionHandler) parseID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("missing %s", param)
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format for %s", param)
	}
	return id, nil
}
