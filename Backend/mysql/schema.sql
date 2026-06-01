CREATE DATABASE IF NOT EXISTS auction_live
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE auction_live;

CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(64) PRIMARY KEY,
  nickname VARCHAR(128) NOT NULL,
  avatar VARCHAR(512) NULL,
  role VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_users_role (role)
);

CREATE TABLE IF NOT EXISTS live_rooms (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  cover_image VARCHAR(512) NULL,
  video_url VARCHAR(512) NULL,
  status VARCHAR(32) NOT NULL,
  anchor_user_id VARCHAR(64) NOT NULL,
  anchor_name VARCHAR(128) NOT NULL,
  online_count INT NOT NULL DEFAULT 0,
  thumbnail VARCHAR(512) NULL,
  current_session_id VARCHAR(64) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_rooms_anchor (anchor_user_id),
  INDEX idx_rooms_status (status)
);

CREATE TABLE IF NOT EXISTS auction_items (
  id VARCHAR(64) PRIMARY KEY,
  room_id VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  cover_image VARCHAR(512) NOT NULL,
  description TEXT NOT NULL,
  start_price BIGINT NOT NULL,
  increment_step BIGINT NOT NULL,
  ceiling_price BIGINT NULL,
  duration_seconds INT NOT NULL,
  extension_seconds INT NOT NULL,
  extension_trigger_seconds INT NOT NULL,
  queue_status VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_items_room (room_id),
  INDEX idx_items_room_queue (room_id, queue_status)
);

CREATE TABLE IF NOT EXISTS room_item_queue (
  room_id VARCHAR(64) NOT NULL,
  item_id VARCHAR(64) NOT NULL,
  sort_order INT NOT NULL,
  PRIMARY KEY (room_id, item_id),
  UNIQUE KEY uniq_room_sort (room_id, sort_order),
  INDEX idx_room_item_queue_room_sort (room_id, sort_order)
);

CREATE TABLE IF NOT EXISTS auction_sessions (
  id VARCHAR(64) PRIMARY KEY,
  room_id VARCHAR(64) NOT NULL,
  item_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  current_price BIGINT NOT NULL,
  leader_user_id VARCHAR(64) NOT NULL DEFAULT '',
  end_time DATETIME NULL,
  participant_count INT NOT NULL DEFAULT 0,
  increment_step BIGINT NOT NULL,
  extension_seconds INT NOT NULL,
  extension_trigger_seconds INT NOT NULL,
  ceiling_price BIGINT NULL,
  supports_auto_proxy TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_sessions_room_status (room_id, status),
  INDEX idx_sessions_item (item_id),
  INDEX idx_sessions_end_time (end_time)
);

CREATE TABLE IF NOT EXISTS bids (
  id VARCHAR(64) PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  room_id VARCHAR(64) NOT NULL,
  item_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  bid_price BIGINT NOT NULL,
  request_id VARCHAR(128) NOT NULL,
  rank_after INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL,
  create_time DATETIME NOT NULL,
  UNIQUE KEY uniq_bids_request_id (request_id),
  INDEX idx_bids_session_time (session_id, create_time),
  INDEX idx_bids_user_time (user_id, create_time)
);

CREATE TABLE IF NOT EXISTS auction_results (
  id VARCHAR(64) PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  item_id VARCHAR(64) NOT NULL,
  result_status VARCHAR(32) NOT NULL,
  winner_user_id VARCHAR(64) NOT NULL DEFAULT '',
  final_price BIGINT NOT NULL DEFAULT 0,
  participant_count INT NOT NULL DEFAULT 0,
  ended_at DATETIME NOT NULL,
  UNIQUE KEY uniq_results_session (session_id),
  INDEX idx_results_item (item_id),
  INDEX idx_results_winner (winner_user_id)
);

CREATE TABLE IF NOT EXISTS auction_orders (
  id VARCHAR(64) PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  room_id VARCHAR(64) NOT NULL,
  item_id VARCHAR(64) NOT NULL,
  buyer_user_id VARCHAR(64) NOT NULL,
  amount BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  create_time DATETIME NOT NULL,
  UNIQUE KEY uniq_orders_session (session_id),
  INDEX idx_orders_buyer_time (buyer_user_id, create_time),
  INDEX idx_orders_room_time (room_id, create_time)
);

CREATE TABLE IF NOT EXISTS room_comments (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  room_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  nickname VARCHAR(128) NOT NULL,
  content VARCHAR(255) NOT NULL,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_room_comments_room_time (room_id, create_time)
);

CREATE TABLE IF NOT EXISTS operation_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL,
  operator_id VARCHAR(64) NULL,
  room_id VARCHAR(64) NULL,
  target_type VARCHAR(64) NOT NULL,
  target_id VARCHAR(64) NOT NULL,
  content TEXT NULL,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_operation_logs_module_time (module, create_time),
  INDEX idx_operation_logs_room_time (room_id, create_time)
);
