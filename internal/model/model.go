package model

import "time"

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       string    `json:"price"`
	ImageURL    string    `json:"imageUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Order struct {
	ID             int       `json:"id"`
	OrderNo        string    `json:"orderNo"`
	ProductID      int       `json:"productId"`
	Amount         string    `json:"amount"`
	ChainCode      string    `json:"chainCode"`
	CoinCode       string    `json:"coinCode"`
	PayOrderID     string    `json:"payOrderId,omitempty"`
	AccessSign     string    `json:"accessSign,omitempty"`
	ReceiptAddress string    `json:"receiptAddress,omitempty"`
	PayURL         string    `json:"payUrl,omitempty"`
	PayTxID        string    `json:"payTxId,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Joined fields (not stored)
	ProductName string `json:"productName,omitempty"`
}

type CoinInfo struct {
	ChainCode       string `json:"chain_code"`
	CoinCode        string `json:"coin_code"`
	ChainLabel      string `json:"chain_label"`
	CoinLabel       string `json:"coin_label"`
	ContractAddress string `json:"contract_address"`
	Decimals        int    `json:"decimals"`
}

type ApiKeyInfo struct {
	ChainCode   string
	Splitter    string
	AccessKeyID string
	SecretKey   string
}
