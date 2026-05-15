package recommendationhandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	recommendationmodel "misis_kolhoz/internal/recommendation/model"
	recommendationservice "misis_kolhoz/internal/recommendation/service"
	"misis_kolhoz/pkg/logger"
)

type mockRecommendationService struct {
	response recommendationmodel.Response
	err      error
	clientID int
}

func (m *mockRecommendationService) GetCachedRecommendations(ctx context.Context, clientID int) (recommendationmodel.Response, error) {
	m.clientID = clientID
	if m.err != nil {
		return recommendationmodel.Response{}, m.err
	}
	return m.response, nil
}

func TestGetRecommendationsStatusOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := logger.NewLogger(context.Background())
	service := &mockRecommendationService{
		response: recommendationmodel.Response{
			ClientID:        1,
			Recommendations: []recommendationmodel.Recommendation{{ProductID: 10, Category: "Tea"}},
		},
	}
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/recommendations/:id", handler.GetRecommendations(ctx))

	req := httptest.NewRequest(http.MethodGet, "/recommendations/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if service.clientID != 1 {
		t.Fatalf("expected client id 1, got %d", service.clientID)
	}
}

func TestGetRecommendationsStatusBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := logger.NewLogger(context.Background())
	handler := NewHandler(&mockRecommendationService{})

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/recommendations/:id", handler.GetRecommendations(ctx))

	req := httptest.NewRequest(http.MethodGet, "/recommendations/not-a-number", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetRecommendationsStatusNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := logger.NewLogger(context.Background())
	handler := NewHandler(&mockRecommendationService{err: recommendationservice.ErrClientNotFound})

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/recommendations/:id", handler.GetRecommendations(ctx))

	req := httptest.NewRequest(http.MethodGet, "/recommendations/2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestGetRecommendationsStatusServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := logger.NewLogger(context.Background())
	handler := NewHandler(&mockRecommendationService{err: recommendationservice.ErrRecommendationsNotReady})

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/recommendations/:id", handler.GetRecommendations(ctx))

	req := httptest.NewRequest(http.MethodGet, "/recommendations/2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}
}

func TestGetRecommendationsStatusInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := logger.NewLogger(context.Background())
	handler := NewHandler(&mockRecommendationService{err: errors.New("boom")})

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/recommendations/:id", handler.GetRecommendations(ctx))

	req := httptest.NewRequest(http.MethodGet, "/recommendations/2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
