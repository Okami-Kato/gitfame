//go:build !solution

package main

import (
	"fmt"
	"os"

	"github.com/Okami-Kato/gitfame/internal/domain"
	"github.com/Okami-Kato/gitfame/internal/engine"
	"github.com/Okami-Kato/gitfame/internal/output"
	"github.com/spf13/pflag"
)

var (
	flagRepository     = pflag.String("repository", "", "Path to the git repository")
	flagRevision       = pflag.String("revision", "HEAD", "Git revision")
	flagUseCommitter   = pflag.Bool("use-committer", false, "If not true - an author will be used to denote a Name")
	flagFormat         = pflag.String("format", "tabular", "Output format option. Can be one of [tabular, csv, json, json-lines]")
	flagOrderBy        = pflag.String("order-by", "lines", "Key that should be used to order the output. Can be one of [lines, commits, files]")
	flagExtensions     = pflag.StringSlice("extensions", nil, "Comma-separated list of white list file extensions")
	flagLanguages      = pflag.StringSlice("languages", nil, "Comma-separated list of white list languages")
	flagExclude        = pflag.StringSlice("exclude", nil, "Comma-separated list of black list glob patterns")
	flagRestrictTo     = pflag.StringSlice("restrict-to", nil, "Comma-separated list of white list glob patterns")
	flagParallelFactor = pflag.Int("parallel-factor", 8, "Amount of routines to run in parallel")
	flagSpinner        = pflag.Bool("spin", false, "Display spinner while waiting for the output")
)

func main() {
	pflag.Parse()
	writer, err := output.NewWriter(output.Format(*flagFormat), os.Stdout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	engine, err := engine.New(&engine.CreationRequest{
		Repository:     *flagRepository,
		Revision:       *flagRevision,
		OrderBy:        *flagOrderBy,
		UseCommitter:   *flagUseCommitter,
		Extensions:     *flagExtensions,
		Languages:      *flagLanguages,
		Exclude:        *flagExclude,
		RestrictTo:     *flagRestrictTo,
		ParallelFactor: *flagParallelFactor,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var fameEntries []domain.FameEntry
	if *flagSpinner {
		fameEntries, err = callAndSpin(engine.Run)
	} else {
		fameEntries, err = engine.Run()
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = writer.Write(fameEntries)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
