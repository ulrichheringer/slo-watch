# slo-watch

Run repeated HTTP checks against one endpoint and report availability, latency objective breaches, and the remaining availability error budget.

## Install

Requires Go 1.23 or newer. Build with `go install github.com/ulrichheringer/slo-watch@latest`.

## Use

Set `url`, `target_availability`, `latency_slo_ms`, `timeout_ms`, and `checks` in a JSON file, then run `slo-watch -config slo.json`.

## Example

Run `go run . -config examples/slo.json`; it checks a local health endpoint and requires that endpoint to be running.
