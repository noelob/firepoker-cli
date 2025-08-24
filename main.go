package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	game := NewGame()
	defer game.Leave()

	err := game.Connect()
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to websocket: %v", err)
		fmt.Println(msg)
		panic(err)
	}

	err = game.Join("d2538816-2f8e-a8b0-6534-30857b5e932d")
	if err != nil {
		msg := fmt.Sprintf("Unable to join the game: %v", err)
		fmt.Println(msg)
		panic(err)
	}

	// Wait until the user closes the application
	wait()
}

func wait() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	//fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
}
