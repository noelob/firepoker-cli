package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	gameId := flag.Arg(0)

	game := NewGame()
	defer game.Leave()

	err := game.Join(gameId)
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

func wait() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	//fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
}
