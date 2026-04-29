package test

import (
    "bytes"
    "fmt"
    "net/http"
    "sync"
    "time"
)

func StressTest() {
    time.Sleep(2 * time.Second) // Даем серверу запуститься
    
    var wg sync.WaitGroup
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            resp, err := http.Post("http://localhost:8082/links", "application/json", 
                bytes.NewBuffer([]byte(`{"url":"https://test.com"}`)))
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            defer resp.Body.Close()
            fmt.Println("Status:", resp.Status)
        }()
    }
    wg.Wait()
}