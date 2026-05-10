package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"misis_kolhoz/internal/vector/model"
	"misis_kolhoz/internal/vector/service"
)

type VectorHandler struct {
	service *service.VectorService
}

func NewVectorHandler(svc *service.VectorService) *VectorHandler {
	return &VectorHandler{service: svc}
}

func (h *VectorHandler) Create(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req model.CreateVectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Vector) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vector is required"})
		return
	}

	err = h.service.Create(c.Request.Context(), id, req.Vector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *VectorHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	vector, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if vector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vector not found"})
		return
	}

	c.JSON(http.StatusOK, model.Vector{ID: id, Vector: vector})
}

func (h *VectorHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req model.CreateVectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.Update(c.Request.Context(), id, req.Vector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *VectorHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *VectorHandler) Search(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ids, err := h.service.Search(c.Request.Context(), req.Vector, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.SearchResponse{IDs: ids})
}

func (h *VectorHandler) Distance(c *gin.Context) {
	var req model.DistanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.VectorA) != len(req.VectorB) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vectors must have same dimensions"})
		return
	}

	cosine, euclidean := h.service.CalculateDistance(req.VectorA, req.VectorB)

	c.JSON(http.StatusOK, model.DistanceResponse{
		Cosine:    cosine,
		Euclidean: euclidean,
	})
}