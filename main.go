package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Config struct {
	URL                string  `json:"url"`
	TargetAvailability float64 `json:"target_availability"`
	LatencySLOMillis   int     `json:"latency_slo_ms"`
	TimeoutMillis      int     `json:"timeout_ms"`
	Checks             int     `json:"checks"`
}

type Result struct {
	Total, Good, LatencyBreaches       int
	Availability, ErrorBudgetRemaining float64
}

func Validate(config Config) error {
	if config.URL == "" {
		return errors.New("config url is required")
	}
	if config.TargetAvailability <= 0 || config.TargetAvailability > 1 {
		return errors.New("target_availability must be between 0 and 1")
	}
	if config.TimeoutMillis <= 0 {
		config.TimeoutMillis = 3000
	}
	if config.Checks <= 0 {
		config.Checks = 10
	}
	return nil
}

func Summarize(total, good int, target float64) Result {
	result := Result{Total: total, Good: good}
	if total == 0 {
		return result
	}
	result.Availability = float64(good) / float64(total)
	if target < 1 {
		result.ErrorBudgetRemaining = (result.Availability - target) / (1 - target)
	}
	if result.ErrorBudgetRemaining < 0 {
		result.ErrorBudgetRemaining = 0
	}
	if result.ErrorBudgetRemaining > 1 {
		result.ErrorBudgetRemaining = 1
	}
	return result
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("slo-watch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to SLO JSON configuration")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return errors.New("usage: slo-watch -config slo.json")
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.Checks <= 0 {
		config.Checks = 10
	}
	if config.TimeoutMillis <= 0 {
		config.TimeoutMillis = 3000
	}
	if err := Validate(config); err != nil {
		return err
	}
	client := &http.Client{Timeout: time.Duration(config.TimeoutMillis) * time.Millisecond}
	good := 0
	latencyBreaches := 0
	for i := 0; i < config.Checks; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, config.URL, nil)
		if err != nil {
			cancel()
			return err
		}
		started := time.Now()
		response, requestErr := client.Do(request)
		elapsed := time.Since(started)
		cancel()
		if requestErr != nil {
			fmt.Fprintf(stderr, "check %d: %v\n", i+1, requestErr)
			continue
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 400 {
			good++
		}
		if config.LatencySLOMillis > 0 && elapsed > time.Duration(config.LatencySLOMillis)*time.Millisecond {
			latencyBreaches++
		}
	}
	result := Summarize(config.Checks, good, config.TargetAvailability)
	result.LatencyBreaches = latencyBreaches
	if _, err := fmt.Fprintf(stdout, "URL: %s\nChecks: %d (%d successful)\nAvailability: %.3f%% (target %.3f%%)\nError budget remaining: %.1f%%\nLatency objective breaches: %d/%d\n", config.URL, result.Total, result.Good, result.Availability*100, config.TargetAvailability*100, result.ErrorBudgetRemaining*100, latencyBreaches, config.Checks); err != nil {
		return err
	}
	if result.Availability < config.TargetAvailability {
		return errors.New("availability SLO missed")
	}
	return nil
}
