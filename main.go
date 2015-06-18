package main

import (
	"fmt"
	"os"
	"os/signal"
)

// main creates the Yeelight driver and starts it in the Ninja Sphere way
func main() {

	NewYeelightDriver()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

}
