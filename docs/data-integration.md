# AKShare + Baostock Data Integration

This project keeps the strategy engine in Go and uses a small Python sidecar for real market data.

## Why this shape

- AKShare has broad coverage across A-shares, ETFs, funds, indices, macro, sector, and fundamental datasets.
- Baostock is stable for A-share daily historical K-line data.
- Both libraries are Python-first, so the sidecar exports normalized CSV files instead of forcing Python packages into the Go runtime.

## Flow

```text
data/universe/ai_tech.csv
        |
        v
scripts/sync_market_data.py
        |
        v
data/live/{universe.csv,bars.csv,fundamentals.csv,market_breadth.csv}
        |
        v
go run ./cmd/rst report -data data/live
```

## Commands

```bash
python3 -m pip install -r requirements-data.txt

go run ./cmd/rst sync-data \
  -provider auto \
  -universe data/universe/ai_tech.csv \
  -out data/live \
  -start 20240101

go run ./cmd/rst report -data data/live
```

## Provider modes

- `auto`: Baostock for stock/index history when suitable, AKShare for ETF/index/fundamental coverage.
- `akshare`: force AKShare.
- `baostock`: force Baostock where supported.

## Output contract

The Go engine reads the same files for fixture, CSV, and real-data runs:

- `universe.csv`
- `bars.csv`
- `fundamentals.csv`
- `market_breadth.csv`

This keeps data acquisition separate from strategy scoring.

## Trading boundary

AKShare and Baostock are market-data libraries. They do not query a broker account's real holdings, fills, cash, or orders. For broker state or execution, add a separate broker adapter such as QMT/XtQuant or PTrade and keep it read-only before any automated order workflow.
