package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

type CSVProvider struct {
	Dir string
}

func (p CSVProvider) Load() (domain.DataSet, error) {
	securities, err := loadUniverse(filepath.Join(p.Dir, "universe.csv"))
	if err != nil {
		return domain.DataSet{}, err
	}
	bars, err := loadBars(filepath.Join(p.Dir, "bars.csv"))
	if err != nil {
		return domain.DataSet{}, err
	}
	fundamentals, err := loadFundamentals(filepath.Join(p.Dir, "fundamentals.csv"))
	if err != nil && !os.IsNotExist(err) {
		return domain.DataSet{}, err
	}
	breadth, err := loadBreadth(filepath.Join(p.Dir, "market_breadth.csv"))
	if err != nil && !os.IsNotExist(err) {
		return domain.DataSet{}, err
	}
	return domain.DataSet{
		Securities:   securities,
		Bars:         bars,
		Fundamentals: fundamentals,
		Breadth:      breadth,
	}, nil
}

func loadUniverse(path string) ([]domain.Security, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	securities := make([]domain.Security, 0, len(rows))
	for _, row := range rows {
		securities = append(securities, domain.Security{
			Symbol:          row["symbol"],
			Name:            row["name"],
			Kind:            domain.AssetKind(row["kind"]),
			Sector:          row["sector"],
			Chain:           row["chain"],
			BenchmarkSymbol: row["benchmark_symbol"],
		})
	}
	return securities, nil
}

func loadBars(path string) (map[string][]domain.Bar, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	bars := make(map[string][]domain.Bar)
	for _, row := range rows {
		date, err := time.Parse("2006-01-02", row["date"])
		if err != nil {
			return nil, fmt.Errorf("parse bar date for %s: %w", row["symbol"], err)
		}
		bar := domain.Bar{
			Symbol:    row["symbol"],
			Date:      date,
			Open:      parseFloat(row["open"]),
			High:      parseFloat(row["high"]),
			Low:       parseFloat(row["low"]),
			Close:     parseFloat(row["close"]),
			Volume:    parseFloat(row["volume"]),
			Amount:    parseFloat(row["amount"]),
			Turnover:  parseFloat(row["turnover"]),
			LimitUp:   parseBool(row["limit_up"]),
			LimitDown: parseBool(row["limit_down"]),
			Paused:    parseBool(row["paused"]),
			AdjFactor: parseFloatDefault(row["adj_factor"], 1),
		}
		bars[bar.Symbol] = append(bars[bar.Symbol], bar)
	}
	return bars, nil
}

func loadFundamentals(path string) (map[string]domain.Fundamental, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	fundamentals := make(map[string]domain.Fundamental)
	for _, row := range rows {
		f := domain.Fundamental{
			Symbol:                 row["symbol"],
			RevenueGrowth:          parseFloat(row["revenue_growth"]),
			ProfitGrowth:           parseFloat(row["profit_growth"]),
			ROE:                    parseFloat(row["roe"]),
			GrossMargin:            parseFloat(row["gross_margin"]),
			OperatingCashflowRatio: parseFloat(row["operating_cashflow_ratio"]),
			RDRatio:                parseFloat(row["rd_ratio"]),
			AIRelevance:            parseFloat(row["ai_relevance"]),
			ST:                     parseBool(row["st"]),
			DelistingRisk:          parseBool(row["delisting_risk"]),
			ListedDays:             parseInt(row["listed_days"]),
			LossDeteriorating:      parseBool(row["loss_deteriorating"]),
			CashflowPoor:           parseBool(row["cashflow_poor"]),
			GoodwillReceivableRisk: parseBool(row["goodwill_receivable_risk"]),
			RecentReportSlowdown:   parseBool(row["recent_report_slowdown"]),
			ConsecutiveLosses:      parseBool(row["consecutive_losses"]),
		}
		fundamentals[f.Symbol] = f
	}
	return fundamentals, nil
}

func loadBreadth(path string) (domain.MarketBreadth, error) {
	rows, err := readCSV(path)
	if err != nil {
		return domain.MarketBreadth{}, err
	}
	if len(rows) == 0 {
		return domain.MarketBreadth{}, nil
	}
	row := rows[len(rows)-1]
	date, _ := time.Parse("2006-01-02", row["date"])
	return domain.MarketBreadth{
		Date:           date,
		Advancers:      parseInt(row["advancers"]),
		Decliners:      parseInt(row["decliners"]),
		LimitDownCount: parseInt(row["limit_down_count"]),
		TotalAmount:    parseFloat(row["total_amount"]),
		AvgAmount20:    parseFloat(row["avg_amount20"]),
		TechReturn60:   parseFloat(row["tech_return60"]),
		BroadReturn60:  parseFloat(row["broad_return60"]),
	}, nil
}

func readCSV(path string) ([]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	rows := make([]map[string]string, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(record) {
				row[header] = strings.TrimSpace(record[i])
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func parseFloatDefault(s string, fallback float64) float64 {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fallback
	}
	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
