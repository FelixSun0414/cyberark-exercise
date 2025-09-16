--
-- Stored procedure: upsert a URL by its 20-byte url_hash and return (id, row_count, code).
-- Requirements: Unique index on (url_hash)
--

DROP PROCEDURE IF EXISTS sp_upsert_shorten_url;

DELIMITER $$

CREATE PROCEDURE sp_upsert_shorten_url (
    IN p_url      TEXT,
    IN p_url_hash BINARY(20)
)
BEGIN
    DECLARE rc INT DEFAULT 0;

    -- Try insert; on duplicate, abort the insert.
    INSERT IGNORE INTO shorten_urls (url_hash, url)
    VALUES (p_url_hash, p_url);

    -- Capture affected rows BEFORE any other.
    --   ROW_COUNT() = 1  -> inserted new row
    --   ROW_COUNT() = 0  -> duplicate key (existing row)
    SET rc = ROW_COUNT();

    -- Return: id (new or existing), and the row_count of the previous INSERT, as well as the code (could be NULL).
    -- Note: the row_count will be used outside to detect if the shorten_url already exist or not.
    SELECT id, rc AS row_count, code
    FROM shorten_urls
    WHERE url_hash = p_url_hash
    LIMIT 1;
END$$

DELIMITER ;
