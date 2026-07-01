package store

import (
	"database/sql"
	"hashnut-demo-shop/internal/model"
	"strings"
	"time"
)

// trimTrailingZeros removes trailing zeros from a decimal string: "0.010000" → "0.01"
func trimTrailingZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// ---- Products ----

func (s *Store) ListProducts() ([]model.Product, error) {
	rows, err := s.db.Query(`SELECT id, name, description, price, image_url, created_at FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Price = trimTrailingZeros(p.Price)
		products = append(products, p)
	}
	return products, nil
}

func (s *Store) GetProduct(id int) (*model.Product, error) {
	var p model.Product
	err := s.db.QueryRow(`SELECT id, name, description, price, image_url FROM products WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL)
	if err != nil {
		return nil, err
	}
	p.Price = trimTrailingZeros(p.Price)
	return &p, nil
}

// ---- Chains ----

func (s *Store) ListSupportedChains() ([]model.CoinInfo, error) {
	rows, err := s.db.Query(
		`SELECT c.block_chain, c.token_symbol, c.chain_label, c.coin_label, c.contract_address, c.decimals
		 FROM t_coin_info c
		 INNER JOIN t_hashnut_api_key k ON c.block_chain = k.block_chain
		 ORDER BY c.block_chain, c.token_symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coins []model.CoinInfo
	for rows.Next() {
		var c model.CoinInfo
		if err := rows.Scan(&c.BlockChain, &c.TokenSymbol, &c.ChainLabel, &c.CoinLabel, &c.ContractAddress, &c.Decimals); err != nil {
			return nil, err
		}
		coins = append(coins, c)
	}
	return coins, nil
}

func (s *Store) GetApiKey(blockChain string) (*model.ApiKeyInfo, error) {
	var a model.ApiKeyInfo
	err := s.db.QueryRow(
		`SELECT block_chain, splitter, access_key_id, secret_key FROM t_hashnut_api_key WHERE block_chain=$1`,
		blockChain,
	).Scan(&a.BlockChain, &a.Splitter, &a.AccessKeyID, &a.SecretKey)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ---- Orders ----

func (s *Store) CreateOrder(o *model.Order) error {
	return s.db.QueryRow(
		`INSERT INTO orders (order_no, product_id, amount, block_chain, token_symbol, pay_order_id, access_sign, receipt_address, pay_url, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at, updated_at`,
		o.OrderNo, o.ProductID, o.Amount, o.BlockChain, o.TokenSymbol,
		o.PayOrderID, o.AccessSign, o.ReceiptAddress, o.PayURL, o.Status,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (s *Store) GetOrder(orderNo string) (*model.Order, error) {
	var o model.Order
	err := s.db.QueryRow(
		`SELECT o.id, o.order_no, o.product_id, o.amount, o.block_chain, o.token_symbol,
		        COALESCE(o.pay_order_id,''), COALESCE(o.access_sign,''),
		        COALESCE(o.receipt_address,''), COALESCE(o.pay_url,''),
		        COALESCE(o.pay_tx_id,''), o.status, o.created_at, o.updated_at,
		        p.name
		 FROM orders o JOIN products p ON o.product_id = p.id
		 WHERE o.order_no = $1`, orderNo,
	).Scan(&o.ID, &o.OrderNo, &o.ProductID, &o.Amount, &o.BlockChain, &o.TokenSymbol,
		&o.PayOrderID, &o.AccessSign, &o.ReceiptAddress, &o.PayURL,
		&o.PayTxID, &o.Status, &o.CreatedAt, &o.UpdatedAt, &o.ProductName)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) UpdateOrderStatus(orderNo, status, payTxID string) error {
	_, err := s.db.Exec(
		`UPDATE orders SET status=$1, pay_tx_id=COALESCE($2, pay_tx_id), updated_at=$3 WHERE order_no=$4`,
		status, payTxID, time.Now(), orderNo,
	)
	return err
}

func (s *Store) FindOrderByPayOrderID(payOrderID string) (*model.Order, error) {
	var o model.Order
	err := s.db.QueryRow(
		`SELECT id, order_no, product_id, amount, block_chain, token_symbol,
		        COALESCE(pay_order_id,''), COALESCE(access_sign,''),
		        COALESCE(receipt_address,''), COALESCE(pay_url,''),
		        COALESCE(pay_tx_id,''), status, created_at, updated_at
		 FROM orders WHERE pay_order_id = $1`, payOrderID,
	).Scan(&o.ID, &o.OrderNo, &o.ProductID, &o.Amount, &o.BlockChain, &o.TokenSymbol,
		&o.PayOrderID, &o.AccessSign, &o.ReceiptAddress, &o.PayURL,
		&o.PayTxID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
