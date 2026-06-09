# right-side-trading

Go MVP for an A-share AI/tech right-side screening and recommendation system.

The first version does not place orders. It scores candidates with:

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

## Commands

- `report`: generate Markdown, CSV, or HTML recommendation output.
- `score`: alias for `report`.
- `sync-data`: provider hook for future data adapters.
- `backtest`: reserved for the next iteration.

## Strategy

The engine avoids bottom-fishing. A candidate needs a supportive market state, top sector strength, established trend, right-side trigger, healthy volume behavior, and a passing quality guardrail to become an active recommendation.
