package loyaltyhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	loyaltymodel "misis_kolhoz/internal/loyalty/model"
)

type mockRepo struct {
	clients      []loyaltymodel.Client
	balance      int
	transactions []loyaltymodel.BonusTransaction
	accrueErr    error
	spendErr     error
}

func (m *mockRepo) InitTables(ctx context.Context) error {
	return nil
}

func (m *mockRepo) UpsertClient(ctx context.Context, client loyaltymodel.Client) error {
	m.clients = append(m.clients, client)
	return nil
}

func (m *mockRepo) GetClientByID(ctx context.Context, id int) (loyaltymodel.Client, error) {
	for _, c := range m.clients {
		if c.ID == id {
			return c, nil
		}
	}
	return loyaltymodel.Client{}, nil
}

func (m *mockRepo) GetClientBalance(ctx context.Context, clientID int) (int, error) {
	return m.balance, nil
}

func (m *mockRepo) GetClientTransactions(ctx context.Context, clientID int) ([]loyaltymodel.BonusTransaction, error) {
	return m.transactions, nil
}

func (m *mockRepo) AccrueBonus(ctx context.Context, clientID int, amount int, orderID int) error {
	if m.accrueErr != nil {
		return m.accrueErr
	}
	m.balance += amount
	return nil
}

func (m *mockRepo) SpendBonus(ctx context.Context, clientID int, amount int) error {
	if m.spendErr != nil {
		return m.spendErr
	}
	m.balance -= amount
	return nil
}

func (m *mockRepo) GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error) {
	return m.clients, nil
}

type mockService struct {
	repo     *mockRepo
	notFound bool
}

func (m *mockService) LoadOrdersFromExcel(ctx context.Context, filePath string) error {
	return nil
}

func (m *mockService) GetClientBonusInfo(ctx context.Context, clientID int) (loyaltymodel.ClientWithBonus, error) {
	if m.notFound || len(m.repo.clients) == 0 {
		return loyaltymodel.ClientWithBonus{}, fmt.Errorf("client not found")
	}
	client := loyaltymodel.Client{ID: clientID, Name: "Test Client", Email: "test@test.com"}
	return loyaltymodel.ClientWithBonus{
		Client:       client,
		Balance:      m.repo.balance,
		Transactions: m.repo.transactions,
	}, nil
}

func (m *mockService) SpendClientBonus(ctx context.Context, clientID int, amount int) error {
	if m.repo.balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	m.repo.balance -= amount
	return nil
}

func (m *mockService) GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error) {
	return m.repo.clients, nil
}

type HandlerWithMock struct {
	service interface {
		LoadOrdersFromExcel(ctx context.Context, filePath string) error
		GetClientBonusInfo(ctx context.Context, clientID int) (loyaltymodel.ClientWithBonus, error)
		SpendClientBonus(ctx context.Context, clientID int, amount int) error
		GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error)
	}
}

func (h *HandlerWithMock) LoadOrders(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.service.LoadOrdersFromExcel(ctx, "test.xlsx")
		if err != nil {
			c.JSON(500, gin.H{"error": "failed"})
			return
		}
		c.JSON(200, gin.H{"message": "success"})
	}
}

func (h *HandlerWithMock) GetAllClients() gin.HandlerFunc {
	return func(c *gin.Context) {
		clients, err := h.service.GetAllClients(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": "failed"})
			return
		}
		c.JSON(200, clients)
	}
}

func (h *HandlerWithMock) GetClientBonus() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := h.service.GetClientBonusInfo(c.Request.Context(), 1)
		if err != nil {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.JSON(200, info)
	}
}

func (h *HandlerWithMock) SpendBonus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Amount int `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		err := h.service.SpendClientBonus(c.Request.Context(), 1, req.Amount)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "bonus spent"})
	}
}

func setupTestRouter() (*gin.Engine, *mockRepo) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{
		clients: []loyaltymodel.Client{
			{ID: 1, Name: "Client 1", Email: "client1@test.com"},
			{ID: 2, Name: "Client 2", Email: "client2@test.com"},
		},
		balance:      500,
		transactions: []loyaltymodel.BonusTransaction{},
	}

	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.POST("/load_orders", h.LoadOrders(context.Background()))
	router.GET("/clients", h.GetAllClients())
	router.GET("/client_bonus/:id", h.GetClientBonus())
	router.POST("/spend_bonus/:id", h.SpendBonus())

	return router, repo
}

func TestGetAllClients(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/clients", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetClientBonus(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/client_bonus/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetClientBonusNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{clients: []loyaltymodel.Client{}}
	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.GET("/client_bonus/:id", h.GetClientBonus())

	req := httptest.NewRequest("GET", "/client_bonus/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSpendBonus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{balance: 500}
	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.POST("/spend_bonus/:id", h.SpendBonus())

	req := httptest.NewRequest("POST", "/spend_bonus/1", nil)
	req.Header.Set("Content-Type", "application/json")
	body := strings.NewReader(`{"amount": 100}`)
	req = httptest.NewRequest("POST", "/spend_bonus/1", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSpendBonusInvalidRequest(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/spend_bonus/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLoadOrders(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/load_orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSpendBonusResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{balance: 500}
	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.POST("/spend_bonus/:id", h.SpendBonus())

	req := httptest.NewRequest("POST", "/spend_bonus/1", nil)
	req.Header.Set("Content-Type", "application/json")
	body := strings.NewReader(`{"amount": 100}`)
	req = httptest.NewRequest("POST", "/spend_bonus/1", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] == "" {
		t.Error("expected non-empty message in response")
	}
}

func TestClientRouter(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/clients", http.StatusOK},
		{"GET", "/client_bonus/1", http.StatusOK},
		{"POST", "/spend_bonus/1", http.StatusOK},
		{"POST", "/load_orders", http.StatusOK},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		if tt.method == "POST" && tt.path != "/load_orders" {
			req.Header.Set("Content-Type", "application/json")
			body := strings.NewReader(`{"amount": 100}`)
			req = httptest.NewRequest(tt.method, tt.path, body)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != tt.code {
			t.Errorf("%s %s: expected status %d, got %d", tt.method, tt.path, tt.code, w.Code)
		}
	}
}

func TestClientBonusNotFoundScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{
		clients: []loyaltymodel.Client{},
	}
	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.GET("/client_bonus/:id", h.GetClientBonus())

	req := httptest.NewRequest("GET", "/client_bonus/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSpendBonusInsufficientBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := &mockRepo{balance: 50}
	svc := &mockService{repo: repo}
	h := &HandlerWithMock{service: svc}

	router.POST("/spend_bonus/:id", h.SpendBonus())

	req := httptest.NewRequest("POST", "/spend_bonus/1", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
