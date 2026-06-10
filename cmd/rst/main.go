package main

import (
	"flag"
	"fmt"
	"os"
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
		fmt.Println("sync-data is a provider hook. The MVP uses local CSV files or the built-in fixture dataset.")
		fmt.Println("Expected files: universe.csv, bars.csv, fundamentals.csv, market_breadth.csv")
		return nil
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
	fmt.Println("  rst sync-data")
	fmt.Println("  rst backtest")
	return nil
}
