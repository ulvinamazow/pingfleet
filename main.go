package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL      string
	Status   int
	Duration time.Duration
	Err      error
	Attempts int
}

type jsonResult struct {
	URL        string `json:"url"`
	Status     int    `json:"status,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Attempts   int    `json:"attempts"`
	Error      string `json:"error,omitempty"`
}

func (r Result) toJSON() jsonResult {
	jr := jsonResult{
		URL:        r.URL,
		Status:     r.Status,
		DurationMs: r.Duration.Microseconds(),
		Attempts:   r.Attempts,
	}

	if r.Err != nil {
		jr.Error = r.Err.Error()
	}
	return jr
}

func checkURL(ctx context.Context, client *http.Client, url string) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{URL: url, Err: err, Duration: time.Since(start)}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Err: err, Duration: time.Since(start)}
	}
	defer resp.Body.Close()

	return Result{
		URL:      url,
		Status:   resp.StatusCode,
		Duration: time.Since(start),
	}
}

func checkWithRetries(ctx context.Context, client *http.Client, url string, maxRetries int) Result {

	var last Result

	for attempt := 1; attempt <= maxRetries; attempt++ {
		last = checkURL(ctx, client, url)
		last.Attempts = attempt

		if last.Err == nil && last.Status < 500 {
			return last
		}

		select {
		case <-ctx.Done():
			return last
		case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
		}
	}
	return last
}

func worker(ctx context.Context, client *http.Client, jobs <-chan string, results chan<- Result, retries int, wg *sync.WaitGroup) {

	defer wg.Done()
	for url := range jobs {
		results <- checkWithRetries(ctx, client, url, retries)
	}
}

func readURLs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func main() {

	var (
		filePath    = flag.String("file", "", "URL siyahisi olan fayl (hər sətirdə bir URL)")
		workerCount = flag.Int("workers", 3, "paralel worker sayi")
		timeout     = flag.Duration("timeout", 5*time.Second, "ümumi timeout (məs. 10s)")
		retries     = flag.Int("retries", 1, "hər URL üçün maksimum cəhd sayi")
		jsonOutput  = flag.Bool("json", false, "nəticəni JSON formatinda çap et")
	)
	flag.Parse()

	var urls []string
	if *filePath != "" {
		var err error
		urls, err = readURLs(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fayl oxuna bilmədi: %v\n", err)
			os.Exit(2)
		}
	} else {
		urls = []string{
			"https://go.dev",
			"https://github.com",
			"https://golang.org",
			"https://httpbin.org/status/404",
			"https://this-domain-almost-certainly-does-not-exist-123.com",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := &http.Client{Timeout: *timeout}

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup
	for i := 0; i < *workerCount; i++ {
		wg.Add(1)
		go worker(ctx, client, jobs, results, *retries, &wg)
	}

	for _, u := range urls {
		jobs <- u
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	exitCode := 0
	var allResults []Result
	for r := range results {
		allResults = append(allResults, r)
		if r.Err != nil || r.Status >= 400 {
			exitCode = 1
		}
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		jsonResuls := make([]jsonResult, 0, len(allResults))
		for _, r := range allResults {
			jsonResuls = append(jsonResuls, r.toJSON())
		}

		if err := encoder.Encode(jsonResuls); err != nil {
			fmt.Fprintf(os.Stderr, "JSON encode error; %v\n", err)
			exitCode = 1
		}
	} else {
		for _, r := range allResults {
			if r.Err != nil {
				fmt.Printf("[ERROR] %-55s -> %v (attempt: %d)\n", r.URL, r.Err, r.Attempts)
				continue
			}
			mark := "OK"
			if r.Status >= 400 {
				mark = "PROBLEM"
			}
			fmt.Printf("[%-7s] %-55s -> %d (%v, attempt: %d)\n", mark, r.URL, r.Status, r.Duration.Round(time.Millisecond), r.Attempts)
		}
	}

	os.Exit(exitCode)
}
