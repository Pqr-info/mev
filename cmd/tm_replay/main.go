package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    root := flag.String("root", "TIME_MACHINE", "Time Machine root")
    date := flag.String("date", "", "Date YYYY-MM-DD")
    flag.Parse()

    if *date == "" {
        fmt.Println("date required")
        os.Exit(1)
    }

    // e.g. TIME_MACHINE/microstructure/2026/07/20
    var year, month, day string
    fmt.Sscanf(*date, "%4s-%2s-%2s", &year, &month, &day)
    base := filepath.Join(*root, "microstructure", year, month, day)

    err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
            return nil
        }
        b, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        var event map[string]any
        if err := json.Unmarshal(b, &event); err != nil {
            return err
        }
        // Here you can feed event into a replay engine, print summary, etc.
        fmt.Println("Replayed Event Timestamp:", event["timestamp"])
        return nil
    })
    if err != nil {
        fmt.Println("replay error:", err)
        os.Exit(1)
    }
}
