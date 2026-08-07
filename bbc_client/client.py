import os
import requests
from urllib.parse import urljoin
from dotenv import load_dotenv

load_dotenv()

class BBCClient:
    """Simple wrapper for the BBC API 2 (RapidAPI)."""
    BASE_URL = "https://bbc-api2.p.rapidapi.com/"

    def __init__(self):
        self.api_key = os.getenv("RAPIDAPI_KEY")
        self.host = os.getenv("RAPIDAPI_HOST", "bbc-api2.p.rapidapi.com")
        if not self.api_key:
            raise RuntimeError("RAPIDAPI_KEY not set in environment")
        self.headers = {
            "X-RapidAPI-Key": self.api_key,
            "X-RapidAPI-Host": self.host,
        }

    def get_news_by_date(self, date: str):
        """Retrieve news for a specific date (format: YYYY‑MM‑DD)."""
        endpoint = urljoin(self.BASE_URL, f"news/{date}")
        resp = requests.get(endpoint, headers=self.headers)
        resp.raise_for_status()
        return resp.json()
