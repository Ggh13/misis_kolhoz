package farmerhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type FarmerWithProducts struct {
	Farmer   interface{} `json:"farmer"`
	Products interface{} `json:"products"`
}

type FarmerService interface {
	LoadDataFromExcel(ctx context.Context, filePath string) error
	GetFarmerByID(ctx context.Context, id int) (FarmerWithProducts, error)
}

type mockFarmerService struct {
	loadDataCalled bool
	getFarmerCalled bool
	farmerID       int
}

func (m *mockFarmerService) LoadDataFromExcel(ctx context.Context, filePath string) error {
	m.loadDataCalled = true
	return nil
}

func (m *mockFarmerService) GetFarmerByID(ctx context.Context, id int) (FarmerWithProducts, error) {
	m.getFarmerCalled = true
	m.farmerID = id
	return FarmerWithProducts{}, nil
}

type HandlerWithInterface struct {
	service FarmerService
}

func (h *HandlerWithInterface) UploadData(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := h.service.LoadDataFromExcel(contx, "internal/moked_data/farmers_sku.xlsx")
		if err != nil {
			ctx.JSON(500, gin.H{"message": "Failed load data"})
			return
		}
		ctx.JSON(200, gin.H{"message": "Successfully loaded data from excel"})
	}
}

func (h *HandlerWithInterface) GetFarmerData(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := 1
		farmer, err := h.service.GetFarmerByID(contx, id)
		if err != nil {
			ctx.JSON(404, gin.H{"message": "Farmer not found"})
			return
		}
		ctx.JSON(200, farmer)
	}
}

func TestUploadData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("upload_data returns 200 on success", func(t *testing.T) {
		handler := &HandlerWithInterface{
			service: &mockFarmerService{},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/upload_data", nil)

		handler.UploadData(context.Background())(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestGetFarmerData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get_farmer_data returns farmer data", func(t *testing.T) {
		mock := &mockFarmerService{}
		handler := &HandlerWithInterface{
			service: mock,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/farmer_data/1", nil)

		handler.GetFarmerData(context.Background())(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if !mock.getFarmerCalled {
			t.Error("expected GetFarmerByID to be called")
		}
	})
}

func TestUploadDataHandler_JSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("upload_data returns success message", func(t *testing.T) {
		handler := &HandlerWithInterface{
			service: &mockFarmerService{},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/upload_data", nil)

		handler.UploadData(context.Background())(c)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response["message"] == "" {
			t.Error("expected non-empty message in response")
		}
	})
}

func TestFarmerRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("routes are registered correctly", func(t *testing.T) {
		router := gin.New()

		handler := &HandlerWithInterface{
			service: &mockFarmerService{},
		}

		router.POST("/upload_data", handler.UploadData(context.Background()))
		router.GET("/farmer_data/:id", handler.GetFarmerData(context.Background()))

		req1 := httptest.NewRequest("POST", "/upload_data", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("upload_data route failed with status %d", w1.Code)
		}

		req2 := httptest.NewRequest("GET", "/farmer_data/1", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("farmer_data route failed with status %d", w2.Code)
		}
	})
}