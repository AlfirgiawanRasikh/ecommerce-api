package category

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	categories, err := h.service.GetAll(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category id",
		})
		return
	}

	category, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get category",
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateCategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	category, err := h.service.Create(
		c.Request.Context(),
		request,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create category",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category id",
		})
		return
	}

	var request UpdateCategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	category, err := h.service.Update(
		c.Request.Context(),
		id,
		request,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to update category",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category id",
		})
		return
	}

	err = h.service.Delete(
		c.Request.Context(),
		id,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete category",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}
