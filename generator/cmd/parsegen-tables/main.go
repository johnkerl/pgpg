package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/johnkerl/pgpg/generator/pkg/parsegen"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-o output.json] input.bnf\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var outputPath string
	var cpuProfilePath string
	var memProfilePath string
	var tracePath string
	flag.StringVar(&outputPath, "o", "", "Output JSON file (default stdout)")
	flag.StringVar(&cpuProfilePath, "cpuprofile", "", "Write CPU profile to file")
	flag.StringVar(&memProfilePath, "memprofile", "", "Write memory profile to file")
	flag.StringVar(&tracePath, "trace", "", "Write execution trace to file")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}
	inputPath := flag.Arg(0)

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	stopProfile, err := startProfiling(cpuProfilePath, memProfilePath, tracePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer stopProfile()

	tables, err := parsegen.GenerateTablesFromEBNFWithSourceName(string(inputBytes), absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	jsonBytes, err := parsegen.EncodeTables(tables)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if outputPath == "" || outputPath == "-" {
		_, _ = os.Stdout.Write(jsonBytes)
		_, _ = os.Stdout.Write([]byte("\n"))
		return
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startProfiling(cpuProfilePath string, memProfilePath string, tracePath string) (func(), error) {
	var cpuFile *os.File
	var memFile *os.File
	var traceFile *os.File

	if cpuProfilePath != "" {
		f, err := os.Create(cpuProfilePath)
		if err != nil {
			return nil, err
		}
		cpuFile = f
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			_ = cpuFile.Close()
			return nil, err
		}
	}

	if tracePath != "" {
		f, err := os.Create(tracePath)
		if err != nil {
			if cpuFile != nil {
				pprof.StopCPUProfile()
				_ = cpuFile.Close()
			}
			return nil, err
		}
		traceFile = f
		if err := trace.Start(traceFile); err != nil {
			_ = traceFile.Close()
			if cpuFile != nil {
				pprof.StopCPUProfile()
				_ = cpuFile.Close()
			}
			return nil, err
		}
	}

	if memProfilePath != "" {
		f, err := os.Create(memProfilePath)
		if err != nil {
			if traceFile != nil {
				trace.Stop()
				_ = traceFile.Close()
			}
			if cpuFile != nil {
				pprof.StopCPUProfile()
				_ = cpuFile.Close()
			}
			return nil, err
		}
		memFile = f
	}

	stop := func() {
		if traceFile != nil {
			trace.Stop()
			_ = traceFile.Close()
		}
		if cpuFile != nil {
			pprof.StopCPUProfile()
			_ = cpuFile.Close()
		}
		if memFile != nil {
			runtime.GC()
			_ = pprof.WriteHeapProfile(memFile)
			_ = memFile.Close()
		}
	}
	return stop, nil
}
