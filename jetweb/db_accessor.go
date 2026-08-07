package jetweb

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
)

type OHLCV struct {
    Date   time.Time
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

// GetOHLCV returns OHLCV rows for a ticker between start (inclusive) and end (exclusive).
func GetOHLCV(dbPath string, ticker string, start, end time.Time) ([]OHLCV, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    defer db.Close()

    query := `SELECT date, open, high, low, close, volume FROM stooq WHERE ticker = ? AND date >= ? AND date < ? ORDER BY date ASC`
    rows, err := db.Query(query, ticker, start.Format("2006-01-02"), end.Format("2006-01-02"))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var data []OHLCV
    for rows.Next() {
        var d string
        var o, h, l, c, v float64
        if err := rows.Scan(&d, &o, &h, &l, &c, &v); err != nil {
            return nil, err
        }
        date, _ := time.Parse("2006-01-02", d)
        data = append(data, OHLCV{Date: date, Open: o, High: h, Low: l, Close: c, Volume: v})
    }
    return data, rows.Err()
}
