import datetime
from ..client import BBCClient

def main():
    client = BBCClient()
    today = datetime.date.today().isoformat()
    news = client.get_news_by_date(today)
    print(f"News for {today}:")
    for item in news.get('items', []):
        title = item.get('title') or item.get('headline')
        print(f"- {title}")

if __name__ == "__main__":
    main()
