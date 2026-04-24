-- tokens 表
CREATE TABLE IF NOT EXISTS tokens (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token           VARCHAR(64)  NOT NULL COMMENT 'token字符串',
    order_id        VARCHAR(128) DEFAULT NULL COMMENT '关联外部订单号',
    max_uses        INT NOT NULL DEFAULT 1 COMMENT '最大使用次数',
    used_count      INT NOT NULL DEFAULT 0 COMMENT '已使用次数',
    expires_at      DATETIME DEFAULT NULL COMMENT '过期时间',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=有效 2=已耗尽 3=已过期 4=已禁用',
    note            VARCHAR(256) DEFAULT '' COMMENT '备注',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_token (token),
    KEY idx_order_id (order_id),
    KEY idx_status (status),
    KEY idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Token表';

-- orders 表
CREATE TABLE IF NOT EXISTS orders (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id        VARCHAR(128) NOT NULL COMMENT '外部订单号',
    product_type    VARCHAR(64)  NOT NULL DEFAULT 'photo_id' COMMENT '产品类型',
    customer_name   VARCHAR(128) DEFAULT '' COMMENT '客户姓名',
    customer_phone  VARCHAR(32)  DEFAULT '' COMMENT '客户手机号',
    customer_email  VARCHAR(128) DEFAULT '' COMMENT '客户邮箱',
    amount          DECIMAL(10,2) DEFAULT 0.00 COMMENT '订单金额',
    token_id        BIGINT UNSIGNED DEFAULT NULL COMMENT '关联token ID',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=已创建 2=token已发送 3=已使用 4=已退款',
    raw_payload     JSON COMMENT '原始webhook payload',
    webhook_ip      VARCHAR(45) COMMENT 'webhook来源IP',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_id (order_id),
    KEY idx_status (status),
    KEY idx_created_at (created_at),
    FOREIGN KEY (token_id) REFERENCES tokens(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单记录表';

-- generation_logs 表
CREATE TABLE IF NOT EXISTS generation_logs (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id        BIGINT UNSIGNED NOT NULL COMMENT '使用的token ID',
    token           VARCHAR(64)  NOT NULL COMMENT 'token字符串',
    spec_name       VARCHAR(64)  DEFAULT '' COMMENT '证件照规格',
    gender          VARCHAR(10)  DEFAULT '' COMMENT '性别',
    source_file     VARCHAR(256) NOT NULL COMMENT '上传原图路径',
    result_file     VARCHAR(256) DEFAULT '' COMMENT '结果图路径',
    dashscope_task_id VARCHAR(128) DEFAULT '' COMMENT '阿里云任务ID',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=处理中 2=成功 3=失败',
    error_message   TEXT COMMENT '失败原因',
    processing_ms   INT DEFAULT 0 COMMENT '处理耗时(毫秒)',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_token_id (token_id),
    KEY idx_token (token),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证件照生成记录表';
