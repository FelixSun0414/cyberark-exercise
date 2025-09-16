CREATE OR REPLACE VIEW shorten_urls_readable AS
SELECT id,
       url,
       url_hash,
       HEX(url_hash) AS url_hash_hex,
       code,
       CONVERT(code USING utf8mb4) AS code_str,
       created_at
FROM shorten_urls;