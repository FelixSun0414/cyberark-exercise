import validators
import storage_sqlite
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from fastapi.responses import JSONResponse, RedirectResponse
from fastapi.middleware.cors import CORSMiddleware
from urllib.parse import urlparse

app = FastAPI()

baseUrl = "http://localhost:5000"

origins = ["http://localhost:3000"]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"], 
    allow_headers=["*"],
)

# Ensure database table is created
storage_sqlite.init_db()

class URLRequest(BaseModel):
    url: str

@app.post("/api/shorten")
def shorten_url(request: URLRequest):
    origin_url = request.url

    print(f"URL: {origin_url}")
    # ensure this URL is in a valid format
    if not validators.url(origin_url):
        print("invalid URL")
        raise HTTPException(status_code=400, detail="invalid URL")
    
    code, exists = storage_sqlite.upsert_url(origin_url)
    print(f"code: {code}, already exists: {exists}")

    return JSONResponse(
        status_code=200 if exists else 201,
        content={ "short_code": code, "short_url": f"{baseUrl}/{code}", "origin_url": origin_url }
    )

@app.get("/{short_code}")
def redirect(short_code: str):
    print(f"short_code: {short_code}")
    origin_url = storage_sqlite.get_by_code(short_code)

    if origin_url == "":
        print(f"cannot find original URL for code: {short_code}")
        raise HTTPException(status_code=404, detail="Short code not found")
    
    print(f"found original URL {origin_url} for code: {short_code}")
    return RedirectResponse(url=origin_url, status_code=302)
    
