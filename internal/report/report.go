package report

import (
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"strings"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/strategy/scoring"
)

type Format string

const (
	Markdown Format = "markdown"
	CSV      Format = "csv"
	HTML     Format = "html"
)

func Write(w io.Writer, result scoring.Result, format Format) error {
	switch format {
	case CSV:
		return writeCSV(w, result)
	case HTML:
		return writeHTML(w, result)
	default:
		return writeMarkdown(w, result)
	}
}

func writeMarkdown(w io.Writer, result scoring.Result) error {
	fmt.Fprintf(w, "# Right-Side AI/Tech Recommendation Report\n\n")
	fmt.Fprintf(w, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Market state: **%s** (score %.1f)\n\n", result.Market.State, result.Market.Score)
	fmt.Fprintf(w, "Market reasons: %s\n\n", strings.Join(result.Market.Reasons, "; "))

	fmt.Fprintln(w, "## Top Sectors")
	fmt.Fprintln(w, "| Rank | Sector | Proxy | Score | Rel20 | Rel60 | Allowed |")
	fmt.Fprintln(w, "|---:|---|---|---:|---:|---:|---|")
	for _, s := range result.Sectors {
		fmt.Fprintf(w, "| %d | %s | %s | %.1f | %.1f%% | %.1f%% | %t |\n",
			s.Rank, s.Sector, s.Name, s.Score, s.RelativeReturn20*100, s.RelativeReturn60*100, s.Allowed)
	}

	fmt.Fprintln(w, "\n## Recommendations")
	fmt.Fprintln(w, "| Grade | Symbol | Name | Chain | Score | Sector | Trend | Volume | Fundamental | Watch Price | Risks |")
	fmt.Fprintln(w, "|---|---|---|---|---:|---:|---|---|---|---:|---|")
	for _, rec := range result.Recommendations {
		fmt.Fprintf(w, "| %s | %s | %s | %s | %.1f | %.1f | %s | %s | %s | %.2f | %s |\n",
			rec.Grade,
			rec.Security.Symbol,
			rec.Security.Name,
			rec.Security.Chain,
			rec.TotalScore,
			rec.Sector.Score,
			rec.Trend.Status,
			rec.Volume.Status,
			rec.Fundamental.Summary,
			rec.ObservationPrice,
			strings.Join(rec.Risks, "; "),
		)
	}
	return nil
}

func writeCSV(w io.Writer, result scoring.Result) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()
	_ = writer.Write([]string{
		"grade", "symbol", "name", "chain", "score", "sector_score", "trend_status",
		"volume_status", "fundamental_summary", "trigger", "watch_price", "risks", "reasons",
	})
	for _, rec := range result.Recommendations {
		_ = writer.Write([]string{
			string(rec.Grade),
			rec.Security.Symbol,
			rec.Security.Name,
			rec.Security.Chain,
			fmt.Sprintf("%.1f", rec.TotalScore),
			fmt.Sprintf("%.1f", rec.Sector.Score),
			rec.Trend.Status,
			rec.Volume.Status,
			rec.Fundamental.Summary,
			string(rec.Trend.Trigger),
			fmt.Sprintf("%.2f", rec.ObservationPrice),
			strings.Join(rec.Risks, "; "),
			strings.Join(rec.Reasons, "; "),
		})
	}
	return writer.Error()
}

func writeHTML(w io.Writer, result scoring.Result) error {
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>Right-Side AI/Tech Report</title>")
	b.WriteString("<style>body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;margin:32px;color:#1f2937}table{border-collapse:collapse;width:100%;margin:16px 0}th,td{border:1px solid #d1d5db;padding:8px;text-align:left}th{background:#f3f4f6}.A{color:#047857;font-weight:700}.B{color:#0369a1;font-weight:700}.Watch{color:#b45309;font-weight:700}.Avoid{color:#b91c1c;font-weight:700}</style>")
	b.WriteString("</head><body>")
	fmt.Fprintf(&b, "<h1>Right-Side AI/Tech Recommendation Report</h1><p>Generated: %s</p>", html.EscapeString(time.Now().Format("2006-01-02 15:04:05")))
	fmt.Fprintf(&b, "<p>Market state: <strong>%s</strong> (score %.1f)</p>", result.Market.State, result.Market.Score)
	b.WriteString("<h2>Recommendations</h2><table><thead><tr><th>Grade</th><th>Symbol</th><th>Name</th><th>Chain</th><th>Score</th><th>Trend</th><th>Volume</th><th>Fundamental</th><th>Risks</th></tr></thead><tbody>")
	for _, rec := range result.Recommendations {
		fmt.Fprintf(&b, "<tr><td class=\"%s\">%s</td><td>%s</td><td>%s</td><td>%s</td><td>%.1f</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			html.EscapeString(string(rec.Grade)),
			html.EscapeString(string(rec.Grade)),
			html.EscapeString(rec.Security.Symbol),
			html.EscapeString(rec.Security.Name),
			html.EscapeString(rec.Security.Chain),
			rec.TotalScore,
			html.EscapeString(rec.Trend.Status),
			html.EscapeString(rec.Volume.Status),
			html.EscapeString(rec.Fundamental.Summary),
			html.EscapeString(strings.Join(rec.Risks, "; ")),
		)
	}
	b.WriteString("</tbody></table></body></html>")
	_, err := io.WriteString(w, b.String())
	return err
}
