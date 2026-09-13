// Command drc-notify lists tenancies due within N days and optionally
// fires D-7 / D-3 / due reminders through a pluggable notifier.
//
//	drc-notify -cases cases.json -within 7
//	drc-notify -cases cases.json -fire -notifier stdout
//	drc-notify -cases cases.json -today 2026-06-15 -fire -notifier file -out reminders.jsonl
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
	"github.com/furyheimdall/deposit-return-clock/notify"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("drc-notify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	casesPath := fs.String("cases", "", "JSON file of cases (array or {\"cases\":[...]}); \"-\" = stdin")
	within := fs.Int("within", 7, "list cases whose deadline is within N calendar days")
	fire := fs.Bool("fire", false, "emit reminders whose fire date is today (D-7 / D-3 / due)")
	notifierKind := fs.String("notifier", "stdout", "stdout | file | webhook")
	outPath := fs.String("out", "", "file path when -notifier=file")
	webhook := fs.String("webhook", "", "URL when -notifier=webhook")
	today := fs.String("today", "", "YYYY-MM-DD override (fake clock)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *casesPath == "" {
		return fmt.Errorf("drc-notify: -cases is required")
	}

	items, err := notify.LoadCasesFile(*casesPath)
	if err != nil {
		return err
	}

	opt := notify.Options{
		Clock:  notify.SystemClock{},
		Within: *within,
		Fire:   *fire,
	}
	if *today != "" {
		t, err := time.Parse("2006-01-02", *today)
		if err != nil {
			return fmt.Errorf("drc-notify: -today: want YYYY-MM-DD")
		}
		opt.Today = clock.DateFromTime(t)
	}
	if *fire {
		n, err := notify.NewNotifier(*notifierKind, *outPath, *webhook)
		if err != nil {
			return err
		}
		opt.Notifier = n
	}

	rep, err := notify.Run(context.Background(), items, opt)
	if err != nil {
		return err
	}
	notify.FormatDue(os.Stdout, rep.Today, rep.Due)
	if *fire {
		fmt.Fprintf(os.Stdout, "fired=%d\n", len(rep.Fired))
	}
	return nil
}
