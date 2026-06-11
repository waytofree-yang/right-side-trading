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

The real-data sync path writes the same contract:

```bash
python3 -m pip install -r requirements-data.txt
go run ./cmd/rst sync-data -provider auto -out data/live -start 20240101
go run ./cmd/rst report -data data/live
```

`sync-data` uses `scripts/sync_market_data.py` to call AKShare and Baostock, then writes:

- `universe.csv`: securities and benchmark mapping
- `bars.csv`: daily OHLCV/amount/turnover data
- `fundamentals.csv`: best-effort quality guardrail fields
- `market_breadth.csv`: derived breadth and tech-vs-broad relative strength
- `sync_metadata.json`: provider, date range, counts, and warnings
