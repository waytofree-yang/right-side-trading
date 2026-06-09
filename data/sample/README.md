# Sample data

The CLI runs with an embedded deterministic fixture when `-data` is omitted:

```bash
go run ./cmd/rst report
```

To use real low-cost local data, provide a directory containing:

- `universe.csv`
- `bars.csv`
- `fundamentals.csv`
- `market_breadth.csv`

Use the headers implemented in `internal/data/csv.go`. The data provider is intentionally replaceable so future TuShare, JoinQuant, Wind, Choice, or broker data adapters can feed the same strategy engine.
