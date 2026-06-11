#!/usr/bin/env python3
"""Sync A-share AI/tech market data into the CSV contract used by the Go engine.

AKShare is used for broad coverage, especially ETF/index spot data and optional
fundamental fields. Baostock is preferred for stable A-share stock history.
The script intentionally writes plain CSV files so the Go strategy engine keeps
working without importing Python packages directly.
"""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import math
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable


BAR_COLUMNS = [
    "symbol",
    "date",
    "open",
    "high",
    "low",
    "close",
    "volume",
    "amount",
    "turnover",
    "limit_up",
    "limit_down",
    "paused",
    "adj_factor",
]

FUNDAMENTAL_COLUMNS = [
    "symbol",
    "revenue_growth",
    "profit_growth",
    "roe",
    "gross_margin",
    "operating_cashflow_ratio",
    "rd_ratio",
    "ai_relevance",
    "st",
    "delisting_risk",
    "listed_days",
    "loss_deteriorating",
    "cashflow_poor",
    "goodwill_receivable_risk",
    "recent_report_slowdown",
    "consecutive_losses",
]

MARKET_BREADTH_COLUMNS = [
    "date",
    "advancers",
    "decliners",
    "limit_down_count",
    "total_amount",
    "avg_amount20",
    "tech_return60",
    "broad_return60",
]


@dataclass
class Security:
    symbol: str
    name: str
    kind: str
    sector: str
    chain: str
    benchmark_symbol: str
    source: str


def main() -> int:
    args = parse_args()
    out_dir = Path(args.out)
    out_dir.mkdir(parents=True, exist_ok=True)

    start = normalize_date(args.start)
    end = normalize_date(args.end or dt.date.today().strftime("%Y%m%d"))
    securities = read_universe(Path(args.universe))

    errors: list[str] = []
    bars: list[dict[str, str]] = []
    fundamentals: list[dict[str, str]] = []

    ak = None
    bs = None
    if args.provider in {"auto", "akshare"}:
        ak = import_optional("akshare")
    if args.provider in {"auto", "baostock"}:
        bs = import_optional("baostock")

    baostock_logged_in = False
    if bs is not None:
        login = bs.login()
        if getattr(login, "error_code", "0") != "0":
            errors.append(f"baostock login failed: {getattr(login, 'error_msg', '')}")
        else:
            baostock_logged_in = True

    try:
        for security in securities:
            try:
                rows = fetch_bars(security, start, end, args.provider, ak, bs, args.adjust)
            except Exception as exc:  # noqa: BLE001 - collect per-symbol sync errors
                errors.append(f"{security.symbol} {security.name}: {exc}")
                continue
            bars.extend(rows)

            if security.kind == "Stock":
                fundamentals.append(fetch_fundamental(security, ak))
    finally:
        if baostock_logged_in and bs is not None:
            bs.logout()

    write_universe(out_dir / "universe.csv", securities)
    write_csv(out_dir / "bars.csv", BAR_COLUMNS, bars)
    write_csv(out_dir / "fundamentals.csv", FUNDAMENTAL_COLUMNS, fundamentals)
    write_csv(out_dir / "market_breadth.csv", MARKET_BREADTH_COLUMNS, [build_market_breadth(bars)])
    write_metadata(out_dir / "sync_metadata.json", args, start, end, securities, bars, errors)

    if errors:
        print("sync completed with warnings:", file=sys.stderr)
        for error in errors:
            print(f"- {error}", file=sys.stderr)
    print(f"synced {len(securities)} securities and {len(bars)} bars into {out_dir}")
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sync market data for right-side strategy analysis")
    parser.add_argument("--provider", choices=["auto", "akshare", "baostock"], default="auto")
    parser.add_argument("--universe", default="data/universe/ai_tech.csv")
    parser.add_argument("--out", default="data/live")
    parser.add_argument("--start", default=one_year_ago())
    parser.add_argument("--end", default="")
    parser.add_argument("--adjust", choices=["qfq", "hfq", "none"], default="qfq")
    return parser.parse_args()


def one_year_ago() -> str:
    return (dt.date.today() - dt.timedelta(days=420)).strftime("%Y%m%d")


def normalize_date(value: str) -> str:
    value = value.strip().replace("-", "")
    if len(value) != 8 or not value.isdigit():
        raise ValueError(f"date must be YYYYMMDD or YYYY-MM-DD: {value}")
    return value


def import_optional(name: str):
    try:
        return __import__(name)
    except ImportError as exc:
        raise SystemExit(
            f"missing Python package {name}; install with: python3 -m pip install -r requirements-data.txt"
        ) from exc


def read_universe(path: Path) -> list[Security]:
    with path.open("r", encoding="utf-8-sig", newline="") as f:
        reader = csv.DictReader(f)
        securities = []
        for row in reader:
            securities.append(
                Security(
                    symbol=row["symbol"].strip(),
                    name=row["name"].strip(),
                    kind=row["kind"].strip(),
                    sector=row["sector"].strip(),
                    chain=row["chain"].strip(),
                    benchmark_symbol=row.get("benchmark_symbol", "").strip(),
                    source=row.get("source", "auto").strip().lower() or "auto",
                )
            )
        return securities


def write_universe(path: Path, securities: Iterable[Security]) -> None:
    columns = ["symbol", "name", "kind", "sector", "chain", "benchmark_symbol"]
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=columns)
        writer.writeheader()
        for security in securities:
            writer.writerow(
                {
                    "symbol": security.symbol,
                    "name": security.name,
                    "kind": security.kind,
                    "sector": security.sector,
                    "chain": security.chain,
                    "benchmark_symbol": security.benchmark_symbol,
                }
            )


def write_csv(path: Path, columns: list[str], rows: list[dict[str, str]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=columns)
        writer.writeheader()
        for row in rows:
            writer.writerow({column: row.get(column, "") for column in columns})


def write_metadata(
    path: Path,
    args: argparse.Namespace,
    start: str,
    end: str,
    securities: list[Security],
    bars: list[dict[str, str]],
    errors: list[str],
) -> None:
    payload = {
        "provider": args.provider,
        "universe": args.universe,
        "start": start,
        "end": end,
        "adjust": args.adjust,
        "security_count": len(securities),
        "bar_count": len(bars),
        "synced_at": dt.datetime.now().isoformat(timespec="seconds"),
        "errors": errors,
    }
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")


def fetch_bars(
    security: Security,
    start: str,
    end: str,
    provider: str,
    ak,
    bs,
    adjust: str,
) -> list[dict[str, str]]:
    source = security.source
    if provider != "auto":
        source = provider

    if source == "baostock" and bs is not None and security.kind in {"Stock", "Index"}:
        return fetch_baostock_bars(security, start, end, bs, adjust)

    if ak is None:
        raise RuntimeError("AKShare is required for this symbol/provider")
    return fetch_akshare_bars(security, start, end, ak, adjust)


def fetch_baostock_bars(security: Security, start: str, end: str, bs, adjust: str) -> list[dict[str, str]]:
    code = to_baostock_code(security)
    fields = "date,code,open,high,low,close,volume,amount,turn,pctChg,isST"
    result = bs.query_history_k_data_plus(
        code,
        fields,
        start_date=format_baostock_date(start),
        end_date=format_baostock_date(end),
        frequency="d",
        adjustflag=baostock_adjust_flag(adjust),
    )
    if getattr(result, "error_code", "0") != "0":
        raise RuntimeError(getattr(result, "error_msg", "baostock query failed"))

    rows: list[dict[str, str]] = []
    while result.next():
        item = dict(zip(result.fields, result.get_row_data()))
        if not item.get("close"):
            continue
        rows.append(
            {
                "symbol": security.symbol,
                "date": item["date"],
                "open": item.get("open", ""),
                "high": item.get("high", ""),
                "low": item.get("low", ""),
                "close": item.get("close", ""),
                "volume": item.get("volume", ""),
                "amount": item.get("amount", ""),
                "turnover": decimal_percent(item.get("turn", "")),
                "limit_up": "false",
                "limit_down": "false",
                "paused": "false",
                "adj_factor": "1",
            }
        )
    return rows


def fetch_akshare_bars(security: Security, start: str, end: str, ak, adjust: str) -> list[dict[str, str]]:
    if security.kind == "ETF":
        df = ak.fund_etf_hist_em(
            symbol=security.symbol,
            period="daily",
            start_date=start,
            end_date=end,
            adjust="" if adjust == "none" else adjust,
        )
    elif security.kind == "Index":
        df = fetch_akshare_index_bars(security, start, end, ak)
    else:
        df = ak.stock_zh_a_hist(
            symbol=security.symbol,
            period="daily",
            start_date=start,
            end_date=end,
            adjust="" if adjust == "none" else adjust,
        )
    return normalize_akshare_bars(security.symbol, df)


def fetch_akshare_index_bars(security: Security, start: str, end: str, ak):
    if hasattr(ak, "stock_zh_index_daily_em"):
        for symbol in [security.symbol, with_market_prefix(security.symbol)]:
            try:
                return ak.stock_zh_index_daily_em(symbol=symbol, start_date=start, end_date=end)
            except Exception:
                continue
    if hasattr(ak, "stock_zh_index_daily"):
        df = ak.stock_zh_index_daily(symbol=with_market_prefix(security.symbol))
        return filter_date_range(df, start, end)
    raise RuntimeError("AKShare index daily API is unavailable in this version")


def normalize_akshare_bars(symbol: str, df) -> list[dict[str, str]]:
    rows: list[dict[str, str]] = []
    if df is None or len(df) == 0:
        return rows
    for _, row in df.iterrows():
        date_value = get_first(row, ["日期", "date", "Date"])
        close = get_first(row, ["收盘", "close", "Close"])
        if is_empty(date_value) or is_empty(close):
            continue
        amount = get_first(row, ["成交额", "amount", "Amount"], "0")
        turnover = get_first(row, ["换手率", "turnover", "Turnover"], "0")
        rows.append(
            {
                "symbol": symbol,
                "date": normalize_output_date(str(date_value)),
                "open": clean_number(get_first(row, ["开盘", "open", "Open"], "")),
                "high": clean_number(get_first(row, ["最高", "high", "High"], "")),
                "low": clean_number(get_first(row, ["最低", "low", "Low"], "")),
                "close": clean_number(close),
                "volume": clean_number(get_first(row, ["成交量", "volume", "Volume"], "0")),
                "amount": clean_number(amount),
                "turnover": decimal_percent(turnover),
                "limit_up": "false",
                "limit_down": "false",
                "paused": "false",
                "adj_factor": "1",
            }
        )
    return rows


def fetch_fundamental(security: Security, ak) -> dict[str, str]:
    base = {column: "" for column in FUNDAMENTAL_COLUMNS}
    base.update(
        {
            "symbol": security.symbol,
            "revenue_growth": "0",
            "profit_growth": "0",
            "roe": "0",
            "gross_margin": "0",
            "operating_cashflow_ratio": "0",
            "rd_ratio": "0",
            "ai_relevance": ai_relevance_guess(security),
            "st": "false",
            "delisting_risk": "false",
            "listed_days": "0",
            "loss_deteriorating": "false",
            "cashflow_poor": "false",
            "goodwill_receivable_risk": "false",
            "recent_report_slowdown": "false",
            "consecutive_losses": "false",
        }
    )
    if ak is None:
        return base
    try:
        df = ak.stock_financial_analysis_indicator(symbol=security.symbol, start_year="2020")
    except Exception:
        return base
    if df is None or len(df) == 0:
        return base
    latest = df.iloc[-1]
    base["roe"] = decimal_percent(get_first(latest, ["净资产收益率(%)", "净资产收益率"], "0"))
    base["gross_margin"] = decimal_percent(get_first(latest, ["销售毛利率(%)", "销售毛利率"], "0"))
    base["operating_cashflow_ratio"] = decimal_percent(
        get_first(latest, ["经营现金流量净额/营业总收入", "经营现金流量净额/营业收入"], "0")
    )
    return base


def build_market_breadth(bars: list[dict[str, str]]) -> dict[str, str]:
    by_symbol: dict[str, list[dict[str, str]]] = {}
    amount_by_date: dict[str, float] = {}
    for row in bars:
        by_symbol.setdefault(row["symbol"], []).append(row)
        amount_by_date[row["date"]] = amount_by_date.get(row["date"], 0.0) + to_float(row.get("amount", "0"))

    latest_rows = []
    for rows in by_symbol.values():
        sorted_rows = sorted(rows, key=lambda item: item["date"])
        if len(sorted_rows) >= 2:
            latest_rows.append((sorted_rows[-2], sorted_rows[-1]))

    advancers = 0
    decliners = 0
    limit_down_count = 0
    total_amount = 0.0
    latest_date = ""
    for prev, curr in latest_rows:
        prev_close = to_float(prev.get("close", "0"))
        close = to_float(curr.get("close", "0"))
        pct = close / prev_close - 1 if prev_close > 0 else 0
        if pct > 0:
            advancers += 1
        elif pct < 0:
            decliners += 1
        if pct <= -0.095:
            limit_down_count += 1
        total_amount += to_float(curr.get("amount", "0"))
        latest_date = max(latest_date, curr.get("date", ""))

    recent_dates = sorted(amount_by_date)[-20:]
    avg_amount20 = sum(amount_by_date[date] for date in recent_dates) / len(recent_dates) if recent_dates else 0
    tech_return60 = basket_return(by_symbol, ["515980", "512760", "512480", "159995", "588000"], 60)
    broad_return60 = basket_return(by_symbol, ["000300", "000905"], 60)
    return {
        "date": latest_date,
        "advancers": str(advancers),
        "decliners": str(decliners),
        "limit_down_count": str(limit_down_count),
        "total_amount": format_float(total_amount),
        "avg_amount20": format_float(avg_amount20),
        "tech_return60": format_float(tech_return60),
        "broad_return60": format_float(broad_return60),
    }


def basket_return(by_symbol: dict[str, list[dict[str, str]]], symbols: list[str], lookback: int) -> float:
    returns = []
    for symbol in symbols:
        rows = sorted(by_symbol.get(symbol, []), key=lambda item: item["date"])
        if len(rows) <= lookback:
            continue
        start = to_float(rows[-1 - lookback].get("close", "0"))
        end = to_float(rows[-1].get("close", "0"))
        if start > 0:
            returns.append(end / start - 1)
    if not returns:
        return 0
    return sum(returns) / len(returns)


def get_first(row, names: list[str], default=""):
    for name in names:
        try:
            value = row[name]
        except Exception:
            continue
        if not is_empty(value):
            return value
    return default


def filter_date_range(df, start: str, end: str):
    if df is None or len(df) == 0:
        return df
    date_column = None
    for candidate in ["日期", "date", "Date"]:
        if candidate in df.columns:
            date_column = candidate
            break
    if date_column is None:
        return df
    start_text = normalize_output_date(start)
    end_text = normalize_output_date(end)
    dates = df[date_column].astype(str).map(normalize_output_date)
    return df[(dates >= start_text) & (dates <= end_text)]


def is_empty(value) -> bool:
    if value is None:
        return True
    if isinstance(value, float) and math.isnan(value):
        return True
    return str(value).strip() == ""


def clean_number(value) -> str:
    if is_empty(value):
        return "0"
    text = str(value).strip().replace(",", "")
    if text.endswith("%"):
        return decimal_percent(text[:-1])
    return text


def decimal_percent(value) -> str:
    if is_empty(value):
        return "0"
    number = to_float(str(value).replace("%", ""))
    if abs(number) > 1:
        number = number / 100
    return format_float(number)


def to_float(value: str) -> float:
    try:
        return float(str(value).strip().replace(",", ""))
    except ValueError:
        return 0.0


def format_float(value: float) -> str:
    return f"{value:.8f}".rstrip("0").rstrip(".") if value else "0"


def normalize_output_date(value: str) -> str:
    text = value.strip().replace("/", "-")
    if len(text) == 8 and text.isdigit():
        return f"{text[:4]}-{text[4:6]}-{text[6:]}"
    return text[:10]


def format_baostock_date(value: str) -> str:
    return f"{value[:4]}-{value[4:6]}-{value[6:]}"


def to_baostock_code(security: Security) -> str:
    symbol = security.symbol
    if symbol.startswith("6") or symbol.startswith("9") or symbol.startswith("000"):
        return f"sh.{symbol}"
    return f"sz.{symbol}"


def with_market_prefix(symbol: str) -> str:
    if symbol.startswith("6") or symbol.startswith("000"):
        return f"sh{symbol}"
    return f"sz{symbol}"


def baostock_adjust_flag(adjust: str) -> str:
    # Baostock: 1 后复权, 2 前复权, 3 不复权。
    if adjust == "hfq":
        return "1"
    if adjust == "none":
        return "3"
    return "2"


def ai_relevance_guess(security: Security) -> str:
    text = f"{security.sector} {security.chain}".lower()
    if "ai" in text:
        return "0.9"
    if "chip" in text or "semiconductor" in text:
        return "0.75"
    if "software" in text or "compute" in text:
        return "0.7"
    return "0.5"


if __name__ == "__main__":
    raise SystemExit(main())
