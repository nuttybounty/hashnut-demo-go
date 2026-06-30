-- ======== Demo Shop Schema ========

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS t_coin_info;
DROP TABLE IF EXISTS t_hashnut_api_key;

-- 支持的链+币种（前端从这里读取可用支付选项）
CREATE TABLE t_coin_info (
    chain_code       VARCHAR(32)  NOT NULL,
    coin_code        VARCHAR(32)  NOT NULL,
    chain_label      VARCHAR(64)  NOT NULL DEFAULT '',
    coin_label       VARCHAR(64)  NOT NULL DEFAULT '',
    contract_address VARCHAR(128) NOT NULL DEFAULT '',
    decimals         INT          NOT NULL DEFAULT 6,
    PRIMARY KEY (chain_code, coin_code)
);

-- 每条链的 splitter 地址和对应的 API 密钥
CREATE TABLE t_hashnut_api_key (
    chain_code    VARCHAR(32)  NOT NULL,
    splitter      VARCHAR(128) NOT NULL,
    access_key_id VARCHAR(128) NOT NULL,
    secret_key    VARCHAR(128) NOT NULL,
    PRIMARY KEY (chain_code)
);

-- 商品表（不绑定具体链/币种，由前端用户选择）
CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description TEXT,
    price       NUMERIC(18,6) NOT NULL,
    image_url   VARCHAR(512),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 订单表（记录用户选择的链+币种）
CREATE TABLE orders (
    id               SERIAL PRIMARY KEY,
    order_no         VARCHAR(64) NOT NULL UNIQUE,
    product_id       INT NOT NULL REFERENCES products(id),
    amount           NUMERIC(18,6) NOT NULL,
    chain_code       VARCHAR(32) NOT NULL,
    coin_code        VARCHAR(32) NOT NULL,
    pay_order_id     VARCHAR(64),
    access_sign      VARCHAR(256),
    receipt_address  VARCHAR(128),
    pay_url          VARCHAR(512),
    pay_tx_id        VARCHAR(128),
    status           VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ======== Seed Data ========

-- 链+币种配置（根据实际部署的链来配置）
INSERT INTO t_coin_info (chain_code, coin_code, chain_label, coin_label, contract_address, decimals) VALUES
    ('erc20', 'usdt', 'Ethereum', 'USDT', lower('0xdAC17F958D2ee523a2206206994597C13D831ec7'), 6),
    ('trc20', 'usdt', 'Tron',     'USDT', 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',6);

-- API Key 配置（每条链一个 splitter + 对应的 API 密钥）
INSERT INTO t_hashnut_api_key (chain_code, splitter, access_key_id, secret_key) VALUES
    ('erc20', '${split address}',  '${access-key-id}', 'secret-key'),
    ('trc20', '${split address}', '${access-key-id}', 'secret-key');

-- 商品数据
INSERT INTO products (name, description, price, image_url) VALUES
    ('T-Shirt',            'A comfortable everyday shirt suitable for casual wear, outdoor activities, and brand promotion.',    0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/QmRMVsxvfjzQMGjrdCoG8Smd75r2WC9cwgejMivugQhREu'),
    ('Sports Water Bottle','A reusable bottle designed to keep you hydrated during workouts, travel, and daily activities.',     0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/QmdPp6SXQ2CatXvTrqjBdf43eRuZ3J5tmLXa1t8cRtWh8E'),
    ('Electric Kettle',    'A convenient kitchen appliance used to quickly boil water for tea, coffee, and other hot drinks.',   0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/QmeDEFcLHd1gUgRWCLDLdu1iG545bdkDd9GoKMefLjMK1H'),
    ('Fitness Tracker',    'A wearable device used to monitor daily activity, exercise, sleep, and other health-related data.', 0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/Qma7zVAJNJpDhJJPQRLBUGPJ2GnwXaCUwWymP8STtGky4L'),
    ('Screwdriver',        'A practical hand tool used for tightening and loosening screws during repairs and assembly work.',   0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/QmVSbyVEfr1tAWUCH99TVfVt9vvoVqRy827KxmoWJaW96S'),
    ('Flashlight',         'A portable light source designed for outdoor activities, emergencies, and use in dark environments.',0.01, 'https://getting-voice-trap.quicknode-ipfs.com/ipfs/QmZmV8Ao8BgaXKNA5A85LAwUCxBbwVH6mJR6GKkiWJTjBq');
