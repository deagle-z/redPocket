CREATE TABLE IF NOT EXISTS `lucky_money` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `sender_id` BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '发送用户ID',
    `amount` NUMERIC(20,3) NULL DEFAULT NULL COMMENT '红包总金额',
    `received` NUMERIC(20,3) NULL DEFAULT 0.000 COMMENT '已领取金额',
    `number` INT UNSIGNED NULL DEFAULT NULL COMMENT '红包个数',
    `lucky` TINYINT UNSIGNED NULL DEFAULT NULL COMMENT '是否随机(0=否,1=是)',
    `thunder` INT UNSIGNED NULL DEFAULT NULL COMMENT '雷号(1-9)',
    `chat_id` VARCHAR(64) NULL DEFAULT NULL COMMENT 'Telegram群组ID',
    `red_list` TEXT NULL DEFAULT NULL COMMENT '红包金额数组(JSON格式)',
    `sender_name` VARCHAR(128) NULL DEFAULT NULL COMMENT '发送者名称',
    `lose_rate` NUMERIC(10,2) NULL DEFAULT NULL COMMENT '中雷倍数(默认1.8)',
    `status` TINYINT UNSIGNED NULL DEFAULT 1 COMMENT '状态(1=正常,2=已领完,3=已过期)',
    `created_at` DATETIME(3) NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_sender_id` (`sender_id`),
    INDEX `idx_chat_id` (`chat_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='红包表';

CREATE TABLE IF NOT EXISTS `lucky_history` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '领取用户ID',
    `lucky_id` BIGINT UNSIGNED NOT NULL COMMENT '红包ID',
    `is_thunder` TINYINT UNSIGNED NULL DEFAULT 0 COMMENT '是否中雷(0=否,1=是)',
    `amount` NUMERIC(20,3) NULL DEFAULT NULL COMMENT '领取金额',
    `lose_money` NUMERIC(20,3) NULL DEFAULT 0.000 COMMENT '损失金额(中雷时赔付)',
    `first_name` VARCHAR(128) NULL DEFAULT NULL COMMENT '用户名',
    `created_at` DATETIME(3) NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_lucky_id` (`lucky_id`),
    INDEX `idx_is_thunder` (`is_thunder`),
    INDEX `idx_created_at` (`created_at`),
    UNIQUE KEY `uk_user_lucky` (`user_id`, `lucky_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='红包领取历史表';

CREATE TABLE IF NOT EXISTS `tg_user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` VARCHAR(64) NULL DEFAULT NULL COMMENT 'Telegram用户名',
    `first_name` VARCHAR(128) NULL DEFAULT NULL COMMENT '用户显示名称',
    `tg_id` BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Telegram用户ID',
    `balance` NUMERIC(20,3) NULL DEFAULT 0.000 COMMENT '账户余额',
    `status` TINYINT UNSIGNED NULL DEFAULT 1 COMMENT '状态(1=正常,0=已离开/禁用)',
    `invite_user` BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '邀请人ID',
    `created_at` DATETIME(3) NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tg_id` (`tg_id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_status` (`status`),
    INDEX `idx_invite_user` (`invite_user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Telegram用户表';