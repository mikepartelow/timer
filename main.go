package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/muesli/termenv"

	_ "embed"
)

//go:embed colorcycle.txt
var colorcycle string

func main() {
	logger := slog.Default()
	output := termenv.NewOutput(os.Stdout)

	duration := MustGetDuration(logger)
	fmt.Println("Starting", duration, "timer.")

	Countdown(duration, output)

	s := output.String("Timer expired at", time.Now().Local().Format(time.Kitchen))
	s = s.Foreground(output.Color("#ff0000"))
	fmt.Println(s)

	_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
}

func TickUntil(finish time.Time, tick func()) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	deadline, cancel := context.WithDeadline(context.Background(), finish)
	defer cancel()

	for done := false; !done; {
		select {
		case <-deadline.Done():
			done = true
			break
		case <-ticker.C:
			tick()
		}
	}
}

func Countdown(duration time.Duration, output *termenv.Output) {
	blinker := NewCycler(":", " ")
	colors := NewCycler(strings.Fields(colorcycle)...)

	output.HideCursor()
	go HandleInterrupt(output)
	defer output.ShowCursor()

	output.SaveCursorPosition()

	finish := time.Now().Add(duration).Add(time.Second)
	TickUntil(finish, func() {
		remaining := time.Until(finish)
		minutes := int(remaining.Minutes())
		seconds := int(remaining.Seconds()) % 60

		s := output.String(fmt.Sprintf("%02d%s%02d", minutes, blinker(), seconds))
		s = s.Foreground(output.Color(colors()))

		fmt.Print(s)
		output.RestoreCursorPosition()
	})
}

func HandleInterrupt(output *termenv.Output) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	output.ShowCursor()
	os.Exit(0)
}

func MustGetDuration(logger *slog.Logger) time.Duration {
	usage := func() {
		fmt.Println("Usage: timer <duration>")
		fmt.Println("timer 2m30s")
		os.Exit(1)
	}

	if len(os.Args) != 2 {
		usage()
	}
	duration := os.Args[1]

	d, err := time.ParseDuration(duration)
	if err != nil {
		logger.Info("time.Parseduration", "error", err)
		usage()
	}

	return d
}

func NewCycler[T any](items ...T) func() T {
	i := len(items) - 1

	return func() T {
		i = (i + 1) % len(items)
		return items[i]
	}
}
