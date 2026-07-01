package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	hashnut "github.com/nuttybounty/hashnut-sdk-go/v4"
	sdkmodel "github.com/nuttybounty/hashnut-sdk-go/v4/model"

	"hashnut-demo-shop/internal/config"
	"hashnut-demo-shop/internal/model"
	"hashnut-demo-shop/internal/store"
)

type Handler struct {
	store  *store.Store
	config *config.HashNutConfig

	// 按 secretKey 缓存 SDK Client
	mu      sync.RWMutex
	clients map[string]*hashnut.Client
}

func New(s *store.Store, cfg *config.HashNutConfig) *Handler {
	return &Handler{
		store:   s,
		config:  cfg,
		clients: make(map[string]*hashnut.Client),
	}
}

func (h *Handler) getClient(secretKey string) *hashnut.Client {
	h.mu.RLock()
	c, ok := h.clients[secretKey]
	h.mu.RUnlock()
	if ok {
		return c
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	// double check
	if c, ok := h.clients[secretKey]; ok {
		return c
	}
	var opts []hashnut.Option
	if h.config.BaseURL != "" {
		opts = append(opts, hashnut.WithBaseURL(h.config.BaseURL))
	}
	c = hashnut.NewClient(secretKey, h.config.TestMode, opts...)
	h.clients[secretKey] = c
	return c
}

// ListProducts GET /api/products
func (h *Handler) ListProducts(c *gin.Context) {
	products, err := h.store.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

// ListChains GET /api/chains
func (h *Handler) ListChains(c *gin.Context) {
	chains, err := h.store.ListSupportedChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": chains})
}

// CreateOrder POST /api/orders
// Body: { "productId": 1, "blockChain": "ETH", "tokenSymbol": "usdt" }
func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		ProductID   int    `json:"productId" binding:"required"`
		BlockChain  string `json:"blockChain" binding:"required"`
		TokenSymbol string `json:"tokenSymbol" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "productId, blockChain, tokenSymbol are required"})
		return
	}

	product, err := h.store.GetProduct(req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	apiKey, err := h.store.GetApiKey(req.BlockChain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain: " + req.BlockChain})
		return
	}

	sdk := h.getClient(apiKey.SecretKey)
	orderNo := uuid.New().String()

	payOrder, err := sdk.CreateOrder(&sdkmodel.CreateOrderRequest{
		AccessKeyID:     apiKey.AccessKeyID,
		MerchantOrderID: orderNo,
		BlockChain:      req.BlockChain,
		TokenSymbol:     req.TokenSymbol,
		Amount:          product.Price,
		SplitterAddress: apiKey.Splitter,
		Subject:         product.Name,
		ExpireDuration:  600,
	})
	if err != nil {
		log.Printf("HashNut CreateOrder failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create payment order failed: " + err.Error()})
		return
	}

	order := &model.Order{
		OrderNo:        orderNo,
		ProductID:      product.ID,
		Amount:         product.Price,
		BlockChain:     req.BlockChain,
		TokenSymbol:    req.TokenSymbol,
		PayOrderID:     payOrder.PayOrderID,
		AccessSign:     payOrder.AccessSign,
		ReceiptAddress: payOrder.ReceiptAddress,
		PayURL:         payOrder.PayURL,
		Status:         "paying",
	}
	if err := h.store.CreateOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save order failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// GetOrder GET /api/orders/:id
func (h *Handler) GetOrder(c *gin.Context) {
	orderNo := c.Param("id")
	order, err := h.store.GetOrder(orderNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// If still paying, query HashNut for latest status
	if order.Status == "paying" && order.PayOrderID != "" {
		apiKey, err := h.store.GetApiKey(order.BlockChain)
		if err == nil {
			sdk := h.getClient(apiKey.SecretKey)
			payOrder, err := sdk.QueryOrder(&sdkmodel.QueryOrderRequest{
				PayOrderID:      order.PayOrderID,
				MerchantOrderID: order.OrderNo,
				AccessSign:      order.AccessSign,
			})
			if err == nil {
				stateInt, _ := payOrder.State.Int64()
				newStatus := mapHashNutState(int(stateInt))
				if newStatus != order.Status {
					_ = h.store.UpdateOrderStatus(order.OrderNo, newStatus, payOrder.PayTxID)
					order.Status = newStatus
					order.PayTxID = payOrder.PayTxID
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func mapHashNutState(state int) string {
	switch state {
	case sdkmodel.OrderStateInit, sdkmodel.OrderStatePaid:
		return "paying"
	case sdkmodel.OrderStateConfirming:
		return "confirming"
	case sdkmodel.OrderStateSuccess:
		return "fund received"
	case sdkmodel.OrderStateFinish:
		return "finish"
	case sdkmodel.OrderStateFailed:
		return "failed"
	case sdkmodel.OrderStateExpired:
		return "expired"
	case sdkmodel.OrderStateCanceled:
		return "canceled"
	default:
		return "paying"
	}
}

// ConfirmPaid POST /api/orders/:id/confirm
// Body: { "payTxId": "..." }
func (h *Handler) ConfirmPaid(c *gin.Context) {
	orderNo := c.Param("id")
	var req struct {
		PayTxID string `json:"payTxId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payTxId is required"})
		return
	}

	order, err := h.store.GetOrder(orderNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	apiKey, err := h.store.GetApiKey(order.BlockChain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "chain config not found: " + order.BlockChain})
		return
	}

	sdk := h.getClient(apiKey.SecretKey)
	if err := sdk.ConfirmPaid(&sdkmodel.ConfirmPaidRequest{
		PayOrderID:      order.PayOrderID,
		MerchantOrderID: order.OrderNo,
		AccessSign:      order.AccessSign,
		PayTxID:         req.PayTxID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm failed: " + err.Error()})
		return
	}

	_ = h.store.UpdateOrderStatus(orderNo, "paying", req.PayTxID)
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
