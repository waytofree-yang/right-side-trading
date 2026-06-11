package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/waytofree-yang/right-side-trading/internal/data"
	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/report"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/scoring"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return usage()
	}
	switch args[1] {
	case "report", "score":
		return runReport(args[2:])
	case "sync-data":
		return runSyncData(args[2:])
	case "backtest":
		fmt.Println("backtest command is reserved for the next iteration. The strategy modules are unit-tested and report-ready.")
		return nil
	case "-h", "--help", "help":
		return usage()
	default:
		return usage()
	}
}

func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	dataDir := fs.String("data", "", "directory with universe.csv, bars.csv, fundamentals.csv, market_breadth.csv")
	format := fs.String("format", "markdown", "report format: markdown, csv, html")
	out := fs.String("out", "", "output file; stdout when empty")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ds, err := loadDataSet(*dataDir)
	if err != nil {
		return err
	}
	engine := scoring.NewEngine()
	result := engine.Score(ds)

	writer := os.Stdout
	if *out != "" {
		file, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	}
	return report.Write(writer, result, report.Format(strings.ToLower(*format)))
}

func runSyncData(args []string) error {
	fs := flag.NewFlagSet("sync-data", flag.ContinueOnError)
	provider := fs.String("provider", "auto", "data provider: auto, akshare, baostock")
	universe := fs.String("universe", "data/universe/ai_tech.csv", "CSV universe definition")
	out := fs.String("out", "data/live", "output directory for generated CSV data")
	start := fs.String("start", "", "start date: YYYYMMDD or YYYY-MM-DD; default is about 420 days ago")
	end := fs.String("end", "", "end date: YYYYMMDD or YYYY-MM-DD; default is today")
	adjust := fs.String("adjust", "qfq", "price adjustment: qfq, hfq, none")
	python := fs.String("python", "python3", "Python executable with akshare/baostock installed")
	script := fs.String("script", "scripts/sync_market_data.py", "market data sync script")
	verbose := fs.Bool("verbose", false, "print Python environment and per-symbol sync details")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cmdArgs := []string{
		*script,
		"--provider", *provider,
		"--universe", *universe,
		"--out", *out,
		"--adjust", *adjust,
	}
	if *start != "" {
		cmdArgs = append(cmdArgs, "--start", *start)
	}
	if *end != "" {
		cmdArgs = append(cmdArgs, "--end", *end)
	}
	if *verbose {
		cmdArgs = append(cmdArgs, "--verbose")
	}
	cmd := exec.Command(*python, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loadDataSet(dir string) (domain.DataSet, error) {
	if dir == "" {
		return data.FixtureProvider{}.Load()
	}
	return data.CSVProvider{Dir: dir}.Load()
}

func usage() error {
	fmt.Println("Usage:")
	fmt.Println("  rst report [-data DIR] [-format markdown|csv|html] [-out FILE]")
	fmt.Println("  rst score  [-data DIR] [-format csv|markdown|html]")
	fmt.Println("  rst sync-data [-provider auto|akshare|baostock] [-universe FILE] [-out DIR] [-start YYYYMMDD] [-end YYYYMMDD] [-verbose]")
	fmt.Println("  rst backtest")
	return nil
}
