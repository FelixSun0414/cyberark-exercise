import string
import random
import validators
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from fastapi.responses import JSONResponse, RedirectResponse
from fastapi.middleware.cors import CORSMiddleware
from urllib.parse import urlparse

app = FastAPI()

# Behold, the in memory storage!
code_to_url = {}
url_to_code = {}

baseUrl = "http://localhost:5000"
codeSources = string.ascii_letters + string.digits

origins = ["http://localhost:3000"]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"], 
    allow_headers=["*"],
)

# a short code with 7 chars (mix of alphabetic charactors and numerals) is able to cover 62^7 URLs.
def generate_code(length: int = 7) -> str:
    return ''.join(random.choices(codeSources, k=length))

class URLRequest(BaseModel):
    url: str

@app.post("/api/shorten")
def shorten_url(request: URLRequest):
    origin_url = request.url
    rsp_status = 200
    code = ""

    print(f"url: {origin_url}")
    if not validators.url(origin_url):
        print("invalid URL")
        raise HTTPException(status_code=400, detail="invalid URL")
    elif origin_url in url_to_code:
        # if this URL already exists, retrieve the shorten code
        code = url_to_code[origin_url]
        print(f"found code: {code}")
    else:
        # this is a new URL, create a unique shorten code
        code = generate_code()
        while code in code_to_url:
            code = generate_code()
        code_to_url[code] = origin_url
        url_to_code[origin_url] = code

        rsp_status = 201
        print(f"new code created: {code}")

    return JSONResponse(
        status_code=rsp_status,
        content={ "short_code": code, "short_url": f"{baseUrl}/{code}", "origin_url": origin_url }
    )

@app.get("/{short_code}")
def redirect(short_code: str):
    print(f"short_code: {short_code}")
    if short_code in code_to_url:
        return RedirectResponse(url=code_to_url[short_code], status_code=302)
    raise HTTPException(status_code=404, detail="Short code not found")
