package storage

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

// HashSHA256_20 returns the leftmost 20 bytes (160 bits) of SHA-256.
// this should be good enough since for 62^8 records, the conflict rate is (1.6 x 10^-20)
// (initial option is to use MD5, but the conflict rate is 7 x 10^-11)
func urlHashSha256_20(url string) [20]byte {
	full := sha256.Sum256([]byte(url)) // 32 bytes
	var out [20]byte
	copy(out[:], full[:20])
	return out
}

// UpsertURL calls the stored procedure sp_upsert_shorten_url(p_url, p_url_hash).
// It returns the row's id, row_count, and the current shorten code (could be empty if NULL).
//   - if row_count = 0, then the <url - code> mapping already exists
//   - if row_count = 1, then a new mapping was inserted
func UpsertURL(dbClient *DbClient, url string) (id uint64, exists bool, code string, err error) {
	urlHash20 := urlHashSha256_20(url)

	ctx, cancel := dbClient.NewRequestContext(0)
	defer cancel()

	// Argument order: (p_url, p_url_hash)
	// Return: (id, row_count, code)
	row := dbClient.DB.QueryRowContext(ctx, `CALL sp_upsert_shorten_url(?, ?)`, url, urlHash20[:])

	var rowCount int
	var codeBytes sql.NullString // VARBINARY(10) contains Base62 ASCII
	if err = row.Scan(&id, &rowCount, &codeBytes); err != nil {
		return 0, false, "", err
	}
	exists = rowCount != 1
	if codeBytes.Valid {
		code = codeBytes.String
	}
	return id, exists, code, nil
}

// SetCodeForId tries to set the given code for the row (id) only if the code is currently NULL.
//
// SQL pattern:
//
//	UPDATE shorten_urls SET code=? WHERE id=? AND code IS NULL;
//	RowsAffected = 1 -> claimed successfully
//	RowsAffected = 0 -> someone else already set it, read the code out and consider this record already exists
func SetCodeForId(dbClient *DbClient, id uint64, code string) (refreshedCode string, claimed bool, err error) {
	ctx, cancel := dbClient.NewRequestContext(0)
	defer cancel()

	// Try to claim (only if code is NULL)
	res, err := dbClient.DB.ExecContext(ctx, `
		UPDATE shorten_urls SET code = ? WHERE id = ? AND code IS NULL
	`, []byte(code), id) // store code as binary-safe
	if err != nil {
		// Handle unique-key violation (UNIQUE(code)) gracefully
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			// Another row already uses this code. This should not happen since the code is generated from the unique id.
			// Fall through to read-back so the caller can see what is stored for this id.
		} else {
			return "", false, err
		}
	}

	aff, _ := res.RowsAffected()
	if aff == 1 {
		// We successfully claimed it.
		return code, true, nil
	}

	// Someone else likely set it already (or duplicate caused the update to fail). Read back current code.
	var got []byte
	if err := dbClient.DB.QueryRowContext(ctx, `
		SELECT code FROM shorten_urls WHERE id = ? LIMIT 1
	`, id).Scan(&got); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, fmt.Errorf("id not found: %d", id)
		}
		return "", false, err
	}
	return string(got), false, nil
}

// GetURLByCode fetches the original URL by its shorten code.
func GetURLByCode(dbClient *DbClient, code string) (string, bool, error) {
	ctx, cancel := dbClient.NewRequestContext(0)
	defer cancel()

	var url string
	// code column is VARBINARY(10); pass []byte to avoid collation surprises
	if err := dbClient.DB.QueryRowContext(ctx, `
		SELECT url FROM shorten_urls WHERE code = ? LIMIT 1
	`, []byte(code)).Scan(&url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return url, true, nil
}
