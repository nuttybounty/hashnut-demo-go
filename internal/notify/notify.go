package notify

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hashnut-demo-shop/internal/store"
)

type NotifyHandler struct {
	store *store.Store
}

func New(s *store.Store) *NotifyHandler {
	return &NotifyHandler{store: s}
}

// HandleNotify POST /api/notify
// HashNut payment backend will POST payment result to this endpoint
// when the order status changes (e.g. paid, failed, expired).
func (h *NotifyHandler) HandleNotify(c *gin.Context) {
	var payload struct {
		PayOrderID      string `json:"payOrderId"`
		MerchantOrderID string `json:"merchantOrderId"`
		State           int    `json:"state"`
		PayTxID         string `json:"payTxId"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Printf("[Notify] payOrderId=%s merchantOrderId=%s state=%d payTxId=%s",
		payload.PayOrderID, payload.MerchantOrderID, payload.State, payload.PayTxID)

	order, err := h.store.FindOrderByPayOrderID(payload.PayOrderID)
	if err != nil {
		log.Printf("[Notify] order not found for payOrderId=%s", payload.PayOrderID)
		// Return success to avoid HashNut retrying
		c.String(http.StatusOK, "success")
		return
	}

	var status string
	switch {
	case payload.State == 3 || payload.State == 4: // SUCCESS or FINISH
		status = "paid"
	case payload.State == -1:
		status = "failed"
	case payload.State == -2:
		status = "expired"
	case payload.State == -3:
		status = "canceled"
	default:
		status = order.Status
	}

	if err := h.store.UpdateOrderStatus(order.OrderNo, status, payload.PayTxID); err != nil {
		log.Printf("[Notify] update order failed: %v", err)
	}

	// Must return "success" to acknowledge
	c.String(http.StatusOK, "success")
}
