package product

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
	page := parsePositiveInt(
		c.DefaultQuery("page", "1"),
		1,
	)

	limit := parsePositiveInt(
		c.DefaultQuery("limit", "20"),
		20,
	)

	if limit > 100 {
		limit = 100
	}

	minPrice, err := ParseFloat(
		c.Query("minPrice"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid minPrice",
		})
		return
	}

	maxPrice, err := ParseFloat(
		c.Query("maxPrice"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid maxPrice",
		})
		return
	}

	query := ProductQuery{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Brand:    c.Query("brand"),
		Size:     c.Query("size"),
		Color:    c.Query("color"),
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Sort:     c.DefaultQuery("sort", "newest"),
		Page:     page,
		Limit:    limit,
	}

	result, err := h.service.GetAll(
		c.Request.Context(),
		query,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Products,
		"pagination": gin.H{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"totalPages":  result.TotalPages,
			"hasNextPage": result.Page < result.TotalPages,
			"hasPrevPage": result.Page > 1,
		},
	})
}

func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.service.GetBySlug(
		c.Request.Context(),
		slug,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func parsePositiveInt(
	value string,
	fallback int,
) int {
	number, err := strconv.Atoi(value)

	if err != nil || number < 1 {
		return fallback
	}

	return number
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	product, err := h.service.Create(
		c.Request.Context(),
		request,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create product",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product id",
		})
		return
	}

	var request UpdateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})
		return
	}

	product, err := h.service.Update(
		c.Request.Context(),
		id,
		request,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to update product",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product id",
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
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete product",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}
