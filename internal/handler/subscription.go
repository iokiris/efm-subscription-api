package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/middleware"

	"github.com/iokiris/efm-subscription-api/internal/model"
	"github.com/iokiris/efm-subscription-api/internal/service"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	svc service.SubscriptionServiceInterface
}

func NewSubscriptionHandler(svc service.SubscriptionServiceInterface) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// RegisterRoutes регистрирует маршруты для управления подписками.
func (h *SubscriptionHandler) RegisterRoutes(r *gin.Engine, authRequired bool) {
	g := r.Group("/subscriptions")
	if authRequired {
		g.Use(middleware.JWTMiddleware())
	}
	{
		g.POST("", h.Create)
		g.PUT(":id", h.Update)
		g.DELETE(":id", h.Delete)
		g.GET(":id", h.Get)
		g.GET("", h.List)
		g.GET("/summary", h.Summary)
	}
}

// Create godoc
// @Summary		Создать подписку
// @Description	Создаёт новую подписку
// @Tags			subscriptions
// @Accept		json
// @Produce		json
// @Param			body	body		model.Subscription	true	"Данные подписки"
// @Success		201		{object}	model.Subscription
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var in model.Subscription
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	if err := h.svc.Create(ctx, &in); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, in)
}

// Update godoc
// @Summary		Обновить подписку
// @Description	Обновляет существующую подписку по ID
// @Tags			subscriptions
// @Accept		json
// @Produce		json
// @Param			id		path		int	true	"ID подписки"
// @Param			body	body		model.Subscription	true	"Данные подписки"
// @Success		200		{object}	model.Subscription
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var in model.Subscription
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in.ID = id

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	if err := h.svc.Update(ctx, &in); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, in)
}

// Delete godoc
// @Summary		Удалить подписку
// @Description	Удаляет подписку по ID.
// @Tags			subscriptions
// @Produce		json
// @Param			id			path	int	true	"ID подписки"
// @Success		204		""
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	// user_id не требуется от клиента
	if err := h.svc.Delete(ctx, id, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Get godoc
// @Summary		Получить подписку
// @Description	Возвращает подписку по ID
// @Tags			subscriptions
// @Produce		json
// @Param			id	path	int	true	"ID подписки"
// @Success		200	{object}	model.Subscription
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router		/subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	sub, err := h.svc.Get(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// List godoc
// @Summary		Список подписок пользователя
// @Description	Возвращает список подписок по user_id
// @Tags			subscriptions
// @Produce		json
// @Param			user_id	query	string	true	"ID пользователя (UUID)"
// @Success		200		{array}		model.Subscription
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	subs, err := h.svc.List(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subs)
}

// Summary godoc
// @Summary		Сумма по подпискам за период
// @Description	Возвращает сумму цен подписок пользователя, пересекающих интервал [from; to]
// @Tags			subscriptions
// @Produce		json
// @Param			user_id			query	string	true	"ID пользователя"
// @Param			service_name	query	string	false	"Имя сервиса"
// @Param			from			query	string	false	"Дата начала (RFC3339, YYYY-MM-DD или MM-YYYY)"
// @Param			to			query	string	false	"Дата конца (RFC3339, YYYY-MM-DD или MM-YYYY)"
// @Success		200		{object}	map[string]int
// @Failure		400		{object}	map[string]string
// @Router		/subscriptions/summary [get]
func (h *SubscriptionHandler) Summary(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	serviceName := c.Query("service_name")
	from := c.Query("from")
	to := c.Query("to")

	ctx, cancel := contextWithTimeout(c, 5*time.Second)
	defer cancel()

	total, err := h.svc.GetSummary(ctx, userID, serviceName, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total})
}

func parseIDParam(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func contextWithTimeout(c *gin.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), d)
}
