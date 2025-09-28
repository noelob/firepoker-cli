package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

const logFile = "fp.log"

func main() {
	log, err := initLogging()
	defer func() {
		if log != nil {
			_ = log.Close()
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Boom! firepoker-cli init failed", "error", err)
			slog.Error("", "stack", string(debug.Stack()))
			fmt.Printf("%s", "Boom!")
			fmt.Printf("%v.\n", err)
		}
	}()

	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	gameId := flag.Arg(0)

	game := NewGame()
	defer game.Leave()

	err = game.Join(gameId)
	if err != nil {
		msg := fmt.Sprintf("Unable to join the game: %v", err)
		fmt.Println(msg)
		panic(err)
	}

	app := buildUi(game)
	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
	}

	// Wait until the user closes the application
	//wait()
}

func initLogging() (*os.File, error) {
	l, err := os.OpenFile(
		logFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600,
	)
	if err != nil {
		fmt.Errorf("log file %q init failed: %w", l, err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	return l, err
}

func wait() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	//fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
}
