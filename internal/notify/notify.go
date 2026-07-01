package notify

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	hashnut "github.com/nuttybounty/hashnut-sdk-go/v4"
	sdkmodel "github.com/nuttybounty/hashnut-sdk-go/v4/model"

	"hashnut-demo-shop/internal/config"
	"hashnut-demo-shop/internal/store"
)

type NotifyHandler struct {
	store   *store.Store
	config  *config.HashNutConfig
	mu      sync.RWMutex
	clients map[string]*hashnut.Client
}

func New(s *store.Store, cfg *config.HashNutConfig) *NotifyHandler {
	return &NotifyHandler{
		store:   s,
		config:  cfg,
		clients: make(map[string]*hashnut.Client),
	}
}

func (h *NotifyHandler) getClient(secretKey string) *hashnut.Client {
	h.mu.RLock()
	c, ok := h.clients[secretKey]
	h.mu.RUnlock()
	if ok {
		return c
	}

	h.mu.Lock()
	defer h.mu.Unlock()
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

// HandleNotify POST /api/notify
//
// HashNut payment backend sends: payOrderId, merchantOrderId, accessSign, state
// Best practice: query the order via SDK to get the actual state, then update local order.
func (h *NotifyHandler) HandleNotify(c *gin.Context) {
	var payload struct {
		PayOrderID      string `json:"payOrderId"`
		MerchantOrderID string `json:"merchantOrderId"`
		AccessSign      string `json:"accessSign"`
		State           int    `json:"state"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Printf("[Notify] payOrderId=%s merchantOrderId=%s state=%d",
		payload.PayOrderID, payload.MerchantOrderID, payload.State)

	// 1. Find local order
	order, err := h.store.FindOrderByPayOrderID(payload.PayOrderID)
	if err != nil {
		log.Printf("[Notify] order not found for payOrderId=%s", payload.PayOrderID)
		c.String(http.StatusOK, "success")
		return
	}

	// Already in terminal status, skip
	if order.Status == "finish" || order.Status == "failed" ||
		order.Status == "expired" || order.Status == "canceled" {
		log.Printf("[Notify] order %s already in terminal status: %s", order.OrderNo, order.Status)
		c.String(http.StatusOK, "success")
		return
	}

	// 2. Query order via SDK to get actual state
	apiKey, err := h.store.GetApiKey(order.BlockChain)
	if err != nil {
		log.Printf("[Notify] no API key for blockChain=%s, skip query", order.BlockChain)
		c.String(http.StatusOK, "success")
		return
	}

	sdk := h.getClient(apiKey.SecretKey)
	payOrder, err := sdk.QueryOrder(&sdkmodel.QueryOrderRequest{
		PayOrderID:      payload.PayOrderID,
		MerchantOrderID: payload.MerchantOrderID,
		AccessSign:      payload.AccessSign,
	})
	if err != nil {
		log.Printf("[Notify] query order failed: %v", err)
		c.String(http.StatusOK, "success")
		return
	}

	// 3. Update local order based on query result
	stateInt, _ := payOrder.State.Int64()
	newStatus := mapState(int(stateInt))
	payTxID := payOrder.PayTxID

	log.Printf("[Notify] order %s query result: state=%d payTxId=%s -> status=%s",
		order.OrderNo, stateInt, payTxID, newStatus)

	if newStatus != order.Status {
		if err := h.store.UpdateOrderStatus(order.OrderNo, newStatus, payTxID); err != nil {
			log.Printf("[Notify] update order failed: %v", err)
		}
	}

	c.String(http.StatusOK, "success")
}

func mapState(state int) string {
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
