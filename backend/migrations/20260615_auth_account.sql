CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  nickname VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  password_update_time BIGINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_auth_identities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  identity_type TINYINT NOT NULL,
  identity_value VARCHAR(128) DEFAULT NULL,
  verify_time BIGINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_identity_type_value (identity_type, identity_value),
  KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_profiles (
  user_id BIGINT UNSIGNED NOT NULL,
  pet_stage VARCHAR(32) NOT NULL,
  interested_pet_types JSON NOT NULL,
  pet_experience VARCHAR(32) NOT NULL DEFAULT '',
  daily_company_time VARCHAR(32) NOT NULL DEFAULT '',
  living_situation VARCHAR(32) NOT NULL DEFAULT '',
  pet_constraints JSON NULL,
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS verification_codes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code_type TINYINT NOT NULL,
  scene VARCHAR(32) NOT NULL,
  target VARCHAR(128) NOT NULL,
  code_hash VARCHAR(255) NOT NULL,
  attempt_count INT NOT NULL DEFAULT 0,
  expire_time BIGINT NOT NULL,
  use_time BIGINT NOT NULL DEFAULT 0,
  ip_hash VARCHAR(128) NOT NULL DEFAULT '',
  user_agent_hash VARCHAR(128) NOT NULL DEFAULT '',
  provider VARCHAR(32) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_target_scene (target, scene),
  KEY idx_expire_time (expire_time),
  KEY idx_ip_scene (ip_hash, scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  refresh_token_hash VARCHAR(255) NOT NULL,
  device_name VARCHAR(128) NOT NULL DEFAULT '',
  user_agent_hash VARCHAR(128) NOT NULL DEFAULT '',
  ip_hash VARCHAR(128) NOT NULL DEFAULT '',
  last_active_time BIGINT NOT NULL,
  expire_time BIGINT NOT NULL,
  revoke_time BIGINT NOT NULL DEFAULT 0,
  revoke_reason VARCHAR(64) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  UNIQUE KEY uk_refresh_token_hash (refresh_token_hash),
  KEY idx_expire_time (expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS login_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  identity_type TINYINT NOT NULL,
  identity_value VARCHAR(128) NOT NULL,
  ip_hash VARCHAR(128) NOT NULL DEFAULT '',
  fail_count INT NOT NULL DEFAULT 0,
  lock_until_time BIGINT NOT NULL DEFAULT 0,
  last_fail_time BIGINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_identity_ip (identity_type, identity_value, ip_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS security_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  detail JSON NULL,
  ip_hash VARCHAR(128) NOT NULL DEFAULT '',
  user_agent_hash VARCHAR(128) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL,
  update_time BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_user_event (user_id, event_type),
  KEY idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
