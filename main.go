package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const defaultSourceURL = "https://raw.githubusercontent.com/groovy-sky/watcher/main/urls.txt"

const pickCount = 3

func main() {
	sourceURL := flag.String("source", defaultSourceURL, "URL of the plain-text file containing one URL per line")
	flag.Parse()

	urls, err := fetchURLs(*sourceURL)
	if err != nil {
		log.Fatalf("Failed to fetch URL list: %v", err)
	}

	if len(urls) < pickCount {
		log.Fatalf("Need at least %d URLs in the source file, got %d", pickCount, len(urls))
	}

	selected := selectRandom(urls, pickCount)
	fmt.Println("Selected URLs to visit:")
	for i, u := range selected {
		fmt.Printf("  %d. %s\n", i+1, u)
	}

	if err := visitURLs(selected); err != nil {
		log.Fatalf("Failed to visit URLs: %v", err)
	}
}

// fetchURLs downloads a plain-text file from sourceURL and returns one URL per
// non-empty, non-comment line.
func fetchURLs(sourceURL string) ([]string, error) {
	resp, err := http.Get(sourceURL) //nolint:gosec // URL is user-supplied intentionally
	if err != nil {
		return nil, fmt.Errorf("http get %q: %w", sourceURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %q", resp.StatusCode, sourceURL)
	}

	var urls []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return urls, nil
}

// selectRandom returns n randomly chosen elements from urls without replacement.
func selectRandom(urls []string, n int) []string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // non-crypto RNG is fine here
	perm := r.Perm(len(urls))
	selected := make([]string, n)
	for i := 0; i < n; i++ {
		selected[i] = urls[perm[i]]
	}
	return selected
}

// visitURLs opens each URL in a headless Chrome browser via chromedp and
// prints the page title.
func visitURLs(urls []string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 120*time.Second)
	defer cancelTimeout()

	for _, u := range urls {
		fmt.Printf("Visiting: %s\n", u)
		var title string
		if err := chromedp.Run(ctx,
			chromedp.Navigate(u),
			chromedp.Title(&title),
		); err != nil {
			log.Printf("Warning: could not visit %s: %v", u, err)
			continue
		}
		fmt.Printf("  Title: %s\n", title)
	}
	return nil
}
