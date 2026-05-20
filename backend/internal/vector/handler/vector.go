package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"misis_kolhoz/internal/vector/model"
	vectorservice "misis_kolhoz/internal/vector/service"
)

type VectorHandler struct {
	svc *vectorservice.VectorService
}

func NewVectorHandler(svc *vectorservice.VectorService) *VectorHandler {
	return &VectorHandler{svc: svc}
}

func (h *VectorHandler) Create(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req model.UpsertVectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Embedding) != 384 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "embedding size must be 384"})
		return
	}

	if err := h.svc.Upsert(c.Request.Context(), productID, req.Embedding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product_id": productID})
}

func (h *VectorHandler) Get(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	embedding, err := h.svc.Get(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if embedding == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vector not found"})
		return
	}

	c.JSON(http.StatusOK, embedding)
}

func (h *VectorHandler) Update(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req model.UpsertVectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Embedding) != 384 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "embedding size must be 384"})
		return
	}

	if err := h.svc.Update(c.Request.Context(), productID, req.Embedding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product_id": productID})
}

func (h *VectorHandler) Delete(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), productID); err != nil {
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

	results, err := h.svc.Search(c.Request.Context(), req.Vector, int(req.Limit), req.FarmerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.SearchResponse{Products: results})
}

func (h *VectorHandler) SearchEventsForProduct(c *gin.Context) {
	var req model.EventForProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.svc.SearchEventsByProduct(c.Request.Context(), req.ProductID, int(req.Limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.EventSearchResponse{Events: results})
}

func (h *VectorHandler) SearchEvents(c *gin.Context) {
	var req model.EventSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.svc.SearchEvents(c.Request.Context(), req.Vector, int(req.Limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.EventSearchResponse{Events: results})
}

func (h *VectorHandler) MatchEventsToProducts(c *gin.Context) {
	var req model.EventsToProductsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.svc.MatchEventsToProducts(c.Request.Context(), req.Limit, req.FarmerID, req.FutureOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.EventsToProductsResponse{Events: results})
}

func (h *VectorHandler) MatchProductsToEvents(c *gin.Context) {
	var req model.ProductsToEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.svc.MatchProductsToEvents(c.Request.Context(), req.Limit, req.FarmerID, req.FutureOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.ProductsToEventsResponse{Products: results})
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

	cosine, euclidean := h.svc.CalculateDistance(req.VectorA, req.VectorB)
	c.JSON(http.StatusOK, model.DistanceResponse{
		Cosine:    cosine,
		Euclidean: euclidean,
	})
}
