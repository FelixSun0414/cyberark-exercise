-- Initializes the database and the shorten_urls table

-- 1) Create database (case-sensitive, UTF-8)
CREATE DATABASE IF NOT EXISTS url_shortener
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_bin;

USE url_shortener;

-- 2) Create the main table: shorten_urls
-- Columns:
--   id         : auto-increment primary key
--   url        : original long URL
--   url_hash   : 20 bytes of the left part of sha256 of URL for deduplication
--   code       : short code (Base62), typically length 5~8; stored as VARBINARY to keep case-sensitivity
--   created_at : creation timestamp
CREATE TABLE IF NOT EXISTS shorten_urls (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    url        TEXT            NOT NULL,
    url_hash   BINARY(20)      NOT NULL,
    code       VARBINARY(10)   NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    -- Enforce URL deduplication
    UNIQUE KEY uk_url_hash (url_hash),
    -- Enforce short-code uniqueness (NULLs allowed so multiple NULLs don't conflict)
    UNIQUE KEY uk_code (code)
) ENGINE=InnoDB;
