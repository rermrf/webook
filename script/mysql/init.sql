-- WeBook 完整数据库初始化脚本
-- 创建所有需要的数据库

CREATE DATABASE IF NOT EXISTS webook;
CREATE DATABASE IF NOT EXISTS webook_intr;
CREATE DATABASE IF NOT EXISTS webook_payment;
CREATE DATABASE IF NOT EXISTS webook_reward;
CREATE DATABASE IF NOT EXISTS webook_account;
CREATE DATABASE IF NOT EXISTS webook_comment;
CREATE DATABASE IF NOT EXISTS webook_follow;
CREATE DATABASE IF NOT EXISTS webook_tag;
CREATE DATABASE IF NOT EXISTS webook_feed;
CREATE DATABASE IF NOT EXISTS webook_notification;
CREATE DATABASE IF NOT EXISTS webook_credit;
CREATE DATABASE IF NOT EXISTS webook_openapi;
CREATE DATABASE IF NOT EXISTS webook_history;

-- 准备 canal 用户 (用于数据同步)
CREATE USER IF NOT EXISTS 'canal'@'%' IDENTIFIED BY 'canal';
GRANT ALL PRIVILEGES ON *.* TO 'canal'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;

-- ============================================================
-- webook 主库 - 用户、文章等核心数据
-- ============================================================
USE webook;

CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `email` varchar(191) DEFAULT NULL,
    `phone` varchar(191) DEFAULT NULL,
    `password` longtext,
    `nick_name` varchar(128) DEFAULT '',
    `about_me` varchar(1024) DEFAULT '',
    `birthday` date DEFAULT NULL,
    `wechat_open_id` varchar(191) DEFAULT NULL,
    `wechat_union_id` varchar(191) DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_users_email` (`email`),
    UNIQUE KEY `uni_users_phone` (`phone`),
    UNIQUE KEY `uni_users_wechat_open_id` (`wechat_open_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `articles` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `title` varchar(1024) DEFAULT NULL,
    `content` mediumblob,
    `author_id` bigint DEFAULT NULL,
    `status` tinyint unsigned DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_articles_author_id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `published_articles` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `title` varchar(1024) DEFAULT NULL,
    `content` mediumblob,
    `author_id` bigint DEFAULT NULL,
    `status` tinyint unsigned DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_published_articles_author_id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `interactives` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `biz_id` bigint DEFAULT NULL,
    `biz` varchar(128) DEFAULT NULL,
    `read_cnt` bigint DEFAULT 0,
    `like_cnt` bigint DEFAULT 0,
    `collect_cnt` bigint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `biz_id_type` (`biz_id`, `biz`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `user_like_bizs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `biz` varchar(128) DEFAULT NULL,
    `biz_id` bigint DEFAULT NULL,
    `uid` bigint DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    `status` tinyint unsigned DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uid_biz_id_type` (`biz`, `biz_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `user_collection_bizs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `cid` bigint DEFAULT NULL,
    `biz_id` bigint DEFAULT NULL,
    `biz` varchar(128) DEFAULT NULL,
    `uid` bigint DEFAULT NULL,
    `status` tinyint unsigned DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `biz_type_id_uid` (`biz_id`, `biz`, `uid`),
    KEY `idx_user_collection_bizs_cid` (`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `async_sms` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `config` text,
    `retry_cnt` int DEFAULT 0,
    `retry_max` int DEFAULT 3,
    `status` tinyint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_async_sms_status_utime` (`status`, `utime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `jobs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `name` varchar(128) DEFAULT NULL,
    `expression` varchar(128) DEFAULT NULL,
    `executor` varchar(128) DEFAULT NULL,
    `cfg` text,
    `status` tinyint DEFAULT 0,
    `version` int DEFAULT 0,
    `next_time` bigint DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_jobs_name` (`name`),
    KEY `idx_jobs_next_time` (`next_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_comment 评论库
-- ============================================================
USE webook_comment;

CREATE TABLE IF NOT EXISTS `comments` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `biz` varchar(128) NOT NULL,
    `biz_id` bigint NOT NULL,
    `content` text NOT NULL,
    `root_id` bigint DEFAULT NULL,
    `pid` bigint DEFAULT NULL,
    `parent_uid` bigint DEFAULT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_comments_biz_biz_id` (`biz`, `biz_id`),
    KEY `idx_comments_root_id` (`root_id`),
    KEY `idx_comments_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_follow 关注库
-- ============================================================
USE webook_follow;

CREATE TABLE IF NOT EXISTS `follow_relations` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `follower` bigint NOT NULL,
    `followee` bigint NOT NULL,
    `status` tinyint DEFAULT 1,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_follower_followee` (`follower`, `followee`),
    KEY `idx_followee` (`followee`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `follow_statistics` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `followers` bigint DEFAULT 0,
    `followees` bigint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_tag 标签库
-- ============================================================
USE webook_tag;

CREATE TABLE IF NOT EXISTS `tags` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `name` varchar(128) NOT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_uid_name` (`uid`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `tag_bizs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `tid` bigint NOT NULL,
    `uid` bigint NOT NULL,
    `biz` varchar(128) NOT NULL,
    `biz_id` bigint NOT NULL,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_tid_biz_biz_id` (`tid`, `biz`, `biz_id`),
    KEY `idx_biz_biz_id` (`biz`, `biz_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_feed Feed 库
-- ============================================================
USE webook_feed;

CREATE TABLE IF NOT EXISTS `feed_pull_events` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `type` varchar(64) NOT NULL,
    `content` text,
    `ctime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid_ctime` (`uid`, `ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `feed_push_events` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `type` varchar(64) NOT NULL,
    `content` text,
    `ctime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid_ctime` (`uid`, `ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_notification 通知库
-- ============================================================
USE webook_notification;

CREATE TABLE IF NOT EXISTS `notifications` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `key_field` varchar(256) NOT NULL DEFAULT '',
    `biz_id` varchar(64) NOT NULL DEFAULT '',
    `channel` tinyint NOT NULL DEFAULT 0,
    `receiver` varchar(256) NOT NULL DEFAULT '',
    `user_id` bigint NOT NULL DEFAULT 0,
    `template_id` varchar(128) NOT NULL DEFAULT '',
    `template_params` json DEFAULT NULL,
    `content` text,
    `status` tinyint NOT NULL DEFAULT 0,
    `strategy` tinyint NOT NULL DEFAULT 1,
    `group_type` tinyint NOT NULL DEFAULT 0,
    `source_id` bigint NOT NULL DEFAULT 0,
    `source_name` varchar(128) NOT NULL DEFAULT '',
    `target_id` bigint NOT NULL DEFAULT 0,
    `target_type` varchar(64) NOT NULL DEFAULT '',
    `target_title` varchar(256) NOT NULL DEFAULT '',
    `is_read` tinyint NOT NULL DEFAULT 0,
    `ctime` bigint NOT NULL DEFAULT 0,
    `utime` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key_channel` (`key_field`, `channel`),
    KEY `idx_user_ctime` (`user_id`, `ctime` DESC),
    KEY `idx_user_group` (`user_id`, `group_type`, `ctime` DESC),
    KEY `idx_user_unread` (`user_id`, `is_read`, `ctime` DESC),
    KEY `idx_receiver_channel` (`receiver`, `channel`, `ctime` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `notification_transactions` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `notification_id` bigint NOT NULL DEFAULT 0,
    `key_field` varchar(256) NOT NULL DEFAULT '',
    `biz_id` varchar(64) NOT NULL DEFAULT '',
    `status` tinyint NOT NULL DEFAULT 0,
    `check_back_timeout_ms` bigint NOT NULL DEFAULT 30000,
    `next_check_time` bigint NOT NULL DEFAULT 0,
    `retry_count` int NOT NULL DEFAULT 0,
    `max_retry` int NOT NULL DEFAULT 5,
    `ctime` bigint NOT NULL DEFAULT 0,
    `utime` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_notification_id` (`notification_id`),
    UNIQUE KEY `uk_key` (`key_field`),
    KEY `idx_status_check` (`status`, `next_check_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `notification_templates` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `template_id` varchar(128) NOT NULL DEFAULT '',
    `channel` tinyint NOT NULL DEFAULT 0,
    `name` varchar(256) NOT NULL DEFAULT '',
    `content` text NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '',
    `status` tinyint NOT NULL DEFAULT 1,
    `sms_sign` varchar(64) NOT NULL DEFAULT '',
    `sms_provider_template_id` varchar(128) NOT NULL DEFAULT '',
    `ctime` bigint NOT NULL DEFAULT 0,
    `utime` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_template_channel` (`template_id`, `channel`),
    KEY `idx_channel_status` (`channel`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_credit 积分库
-- ============================================================
USE webook_credit;

CREATE TABLE IF NOT EXISTS `credit_accounts` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `balance` bigint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `credit_flows` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `amount` bigint NOT NULL,
    `type` tinyint NOT NULL COMMENT '1=收入 2=支出',
    `biz` varchar(64) DEFAULT '',
    `biz_id` bigint DEFAULT NULL,
    `description` varchar(256) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid_ctime` (`uid`, `ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `credit_recharges` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `amount` bigint NOT NULL,
    `status` tinyint DEFAULT 0 COMMENT '0=待支付 1=成功 2=失败',
    `payment_sn` varchar(128) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid_status` (`uid`, `status`),
    KEY `idx_payment_sn` (`payment_sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `credit_rewards` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `target_uid` bigint NOT NULL,
    `biz` varchar(64) NOT NULL,
    `biz_id` bigint NOT NULL,
    `amount` bigint NOT NULL,
    `status` tinyint DEFAULT 0 COMMENT '0=处理中 1=成功 2=失败',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid` (`uid`),
    KEY `idx_target_uid` (`target_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `daily_credit_records` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `date` date NOT NULL,
    `biz` varchar(64) NOT NULL,
    `earned` bigint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_uid_date_biz` (`uid`, `date`, `biz`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_payment 支付库
-- ============================================================
USE webook_payment;

CREATE TABLE IF NOT EXISTS `payments` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `biz_trade_no` varchar(128) NOT NULL,
    `description` varchar(256) DEFAULT '',
    `total_amount` bigint NOT NULL,
    `status` tinyint DEFAULT 0,
    `txn_id` varchar(128) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_biz_trade_no` (`biz_trade_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_reward 打赏库
-- ============================================================
USE webook_reward;

CREATE TABLE IF NOT EXISTS `rewards` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `target_uid` bigint NOT NULL,
    `biz` varchar(64) NOT NULL,
    `biz_id` bigint NOT NULL,
    `biz_name` varchar(256) DEFAULT '',
    `amount` bigint NOT NULL,
    `status` tinyint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_uid` (`uid`),
    KEY `idx_target_uid` (`target_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_account 账户库
-- ============================================================
USE webook_account;

CREATE TABLE IF NOT EXISTS `accounts` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `uid` bigint NOT NULL,
    `account` bigint DEFAULT 0,
    `type` tinyint DEFAULT 1,
    `currency` varchar(16) DEFAULT 'CNY',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_uid_type_currency` (`uid`, `type`, `currency`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `account_activities` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `aid` bigint NOT NULL,
    `biz` varchar(64) NOT NULL,
    `biz_id` bigint NOT NULL,
    `amount` bigint NOT NULL,
    `ctime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_aid` (`aid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_intr 互动库 (备用)
-- ============================================================
USE webook_intr;

CREATE TABLE IF NOT EXISTS `interactives` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `biz_id` bigint DEFAULT NULL,
    `biz` varchar(128) DEFAULT NULL,
    `read_cnt` bigint DEFAULT 0,
    `like_cnt` bigint DEFAULT 0,
    `collect_cnt` bigint DEFAULT 0,
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `biz_id_type` (`biz_id`, `biz`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_openapi 开放平台库
-- ============================================================
USE webook_openapi;

CREATE TABLE IF NOT EXISTS `apps` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `app_id` varchar(64) NOT NULL,
    `app_secret` varchar(128) NOT NULL,
    `name` varchar(128) NOT NULL,
    `description` varchar(512) DEFAULT '',
    `owner_uid` bigint NOT NULL,
    `type` tinyint DEFAULT 1 COMMENT '1=OAuth2 2=API 3=Both',
    `status` tinyint DEFAULT 0 COMMENT '0=pending 1=approved 2=rejected 3=disabled',
    `redirect_uris` text,
    `scopes` varchar(512) DEFAULT '',
    `ip_whitelist` varchar(512) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_app_id` (`app_id`),
    KEY `idx_owner_uid` (`owner_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `authorizations` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `app_id` varchar(64) NOT NULL,
    `uid` bigint NOT NULL,
    `scope` varchar(256) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    `utime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_app_id_uid` (`app_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `api_call_logs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `app_id` varchar(64) NOT NULL,
    `endpoint` varchar(256) NOT NULL,
    `method` varchar(16) NOT NULL,
    `status_code` int DEFAULT NULL,
    `latency_ms` int DEFAULT NULL,
    `ip` varchar(64) DEFAULT '',
    `ctime` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_app_id_ctime` (`app_id`, `ctime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================
-- webook_history 浏览历史库
-- ============================================================
USE webook_history;

CREATE TABLE IF NOT EXISTS `browse_histories` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `user_id` bigint NOT NULL,
    `biz` varchar(64) NOT NULL DEFAULT 'article',
    `biz_id` bigint NOT NULL,
    `biz_title` varchar(256) NOT NULL DEFAULT '',
    `author_name` varchar(128) NOT NULL DEFAULT '',
    `ctime` bigint NOT NULL DEFAULT 0,
    `utime` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_biz` (`user_id`, `biz`, `biz_id`),
    KEY `idx_user_utime` (`user_id`, `utime` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
