# slo-watch

Send a batch of HTTP checks to an endpoint and compare its availability with your target. The report also shows latency breaches and the remaining error budget.

## Install

With Go 1.23 or newer, install it with `go install github.com/ulrichheringer/slo-watch@latest`.

## Use

Add `url`, `target_availability`, `latency_slo_ms`, `timeout_ms`, and `checks` to a JSON config file, then run `slo-watch -config slo.json`.

## Example

The example checks a local health endpoint, so start that service first. Then run `go run . -config examples/slo.json`.
