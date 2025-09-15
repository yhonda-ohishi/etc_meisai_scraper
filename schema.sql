-- ETC meisai table schema
CREATE TABLE IF NOT EXISTS etc_meisai (
    id INT AUTO_INCREMENT PRIMARY KEY,
    unko_no VARCHAR(50) NOT NULL COMMENT '運行NO',
    date DATE NOT NULL COMMENT '日付',
    time VARCHAR(10) NOT NULL COMMENT '時刻',
    ic_entry VARCHAR(100) COMMENT 'IC入口',
    ic_exit VARCHAR(100) COMMENT 'IC出口',
    vehicle_no VARCHAR(50) NOT NULL COMMENT '車両番号',
    card_no VARCHAR(50) NOT NULL COMMENT 'ETCカード番号',
    amount INT NOT NULL DEFAULT 0 COMMENT '利用金額',
    discount_amount INT NOT NULL DEFAULT 0 COMMENT '割引金額',
    total_amount INT NOT NULL DEFAULT 0 COMMENT '請求金額',
    usage_type VARCHAR(50) COMMENT '利用区分',
    payment_method VARCHAR(50) COMMENT '支払方法',
    route_code VARCHAR(50) COMMENT '路線コード',
    distance DECIMAL(10,2) DEFAULT 0 COMMENT '走行距離',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_unko_no (unko_no),
    INDEX idx_date (date),
    INDEX idx_vehicle_no (vehicle_no),
    INDEX idx_card_no (card_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;