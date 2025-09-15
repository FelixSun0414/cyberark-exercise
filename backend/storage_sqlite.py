import os
import sqlite3
import string
import random
import hashlib
from contextlib import closing

DB_PATH = os.environ.get("DB_PATH", "./db/shorten_urls.db")

# in total 62 ascii chars as the shorten code source
CODE_SOURCES = string.ascii_letters + string.digits

# ----------------------------
# Utils
# ----------------------------
def url_sha256_hex(url: str) -> str:
    """
    Get hash string for the original url
    """
    return hashlib.sha256(url.encode("utf-8")).hexdigest()

def gen_code(length: int = 7) -> str:
    """
    Generate a short code with 7 chars (mix of alphabetic charactors and numerals),
    this should be good enough to cover 62^7 URLs.
    """
    return ''.join(random.choices(CODE_SOURCES, k=length))

# ----------------------------
# Initialize Sqlite database
# ----------------------------
def get_conn():
    conn = sqlite3.connect(DB_PATH, check_same_thread=False)
    conn.execute("PRAGMA journal_mode = WAL;")
    conn.execute("PRAGMA synchronous = NORMAL;")
    conn.execute("PRAGMA foreign_keys = ON;")
    return conn

def init_db():
    """
    Initialize database by creating the required schema if not exists yet
    """
    with closing(get_conn()) as conn, conn:
        conn.execute("""
        CREATE TABLE IF NOT EXISTS shorten_urls (
            id        INTEGER PRIMARY KEY AUTOINCREMENT,
            url       TEXT NOT NULL,
            url_hash  TEXT NOT NULL UNIQUE,
            code      TEXT NOT NULL UNIQUE,
            created_at TIMESTAMP DEFAULT (strftime('%Y-%m-%d %H:%M:%f','now')),
            CHECK (length(url) <= 2048),
            CHECK (length(code) <= 10)
        );
        """)

        conn.execute("CREATE UNIQUE INDEX IF NOT EXISTS idx_shorten_urls_url_hash ON shorten_urls(url_hash);")
        conn.execute("CREATE UNIQUE INDEX IF NOT EXISTS idx_shorten_urls_code ON shorten_urls(code);")

# ----------------------------
# CRUD helpers
# ----------------------------
def upsert_url(url: str):
    """
    Insert a new URL record into the database:
      - if the URL already exists, return the existing shorten code,
      - otherwise, generate a new shorten code for the URL and insert into the database.
    
    Return the shorten code for the original URL and its existence.
    """
    url_hash = url_sha256_hex(url)
    with closing(get_conn()) as conn, conn:
        # already exists?
        row = conn.execute(
            "SELECT code FROM shorten_urls WHERE url_hash = ?",
            (url_hash,)
        ).fetchone()
        if row:
            return row[0], True

        # shorten code does not exist yet, generate a new one and insert into the database
        code = _gen_unique_code(conn)
        conn.execute(
            "INSERT INTO shorten_urls (url, url_hash, code) VALUES (?, ?, ?)",
            (url, url_hash, code)
        )
        return code, False

def get_by_code(code: str) -> str:
    """
    Query original URL by shorten code, return the original URL if found, otherwise return empty string.
    """
    with closing(get_conn()) as conn:
        row = conn.execute(
            "SELECT url FROM shorten_urls WHERE code = ?",
            (code,)
        ).fetchone()
        return row[0] if row else ""

def _gen_unique_code(conn: sqlite3.Connection) -> str:
    """
    Generate a new unique code
    """
    # TODO: better to have max try, otherwise this could take long to find a valid code if there are already lots of records exist
    while True:
        candidate = gen_code()
        exists = conn.execute("SELECT 1 FROM shorten_urls WHERE code = ?", (candidate,)).fetchone()
        if not exists:
            return candidate