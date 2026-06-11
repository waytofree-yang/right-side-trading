# right-side-trading

Go MVP for an A-share AI/tech right-side screening and recommendation system.

The first version does not place orders. It syncs market data, scores candidates, and generates local recommendation reports with:

- market risk gate
- sector relative strength
- individual trend and right-side trigger
- public volume/price capital proxy
- fundamental quality guardrail

## Run

```bash
go run ./cmd/rst report
go run ./cmd/rst report -format csv
go run ./cmd/rst report -format html -out report.html
```

Without `-data`, the CLI uses an embedded deterministic fixture so the system works immediately. With `-data DIR`, it expects local CSV files described in `data/sample/README.md`.

## Real Data

AKShare and Baostock are integrated through a Python sidecar that writes the same CSV contract consumed by the Go strategy engine.

Install optional data dependencies:

```bash
python3 -m pip install -r requirements-data.txt
```

Sync real A-share AI/tech data:

```bash
go run ./cmd/rst sync-data \
  -provider auto \
  -universe data/universe/ai_tech.csv \
  -out data/live \
  -start 20240101
```

Analyze the synced data:

```bash
go run ./cmd/rst report -data data/live
```

Provider behavior:

- `auto`: use Baostock for A-share stock/index history when possible, AKShare for ETF/index/fundamental coverage.
- `akshare`: force AKShare for all supported symbols.
- `baostock`: force Baostock for supported A-share stock/index history.

## Commands

- `report`: generate Markdown, CSV, or HTML recommendation output.
- `score`: alias for `report`.
- `sync-data`: sync AKShare/Baostock data into local CSV files.
- `backtest`: reserved for the next iteration.

## Strategy

The engine avoids bottom-fishing. A candidate needs a supportive market state, top sector strength, established trend, right-side trigger, healthy volume behavior, and a passing quality guardrail to become an active recommendation.

## Boundaries

AKShare and Baostock provide market, historical, and fundamental data. They do not provide broker account execution, holdings, or fills. Broker integration should be handled separately through QMT, PTrade, or a broker-supported API.
