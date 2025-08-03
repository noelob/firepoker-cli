package main

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func wait() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	//fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
}

func main() {
	appCtx, appCtxCancel := context.WithCancel(context.Background())
	defer appCtxCancel()

	c := connect(appCtx)
	defer c.CloseNow()

	go func() {
		// Allows this go routine to read indefinitely until the application cancels its context (i.e. shuts down)
		for {
			ctx, cancel := context.WithCancel(appCtx)
			//fmt.Println("Reading response(s) (1)")
			_, bytes, err := c.Read(ctx)
			if err != nil {
				msg := fmt.Sprintf("<<< Unable to read from websocket: %v", err)
				fmt.Println(msg)
				cancel()
				break
			}
			println("<<< " + string(bytes))
			cancel()
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			keepalive(c)
		}
	}()

	join(c, appCtx)

	// Wait until the user closes the application
	wait()

	// This should cancel all derived contexts (i.e. )
	appCtxCancel()

	fmt.Println("Closing websocket (using handshake)")
	c.Close(websocket.StatusNormalClosure, "")
}

func join(c *websocket.Conn, appCtx context.Context) {
	//fmt.Println("Joining game (1)")
	send(c, appCtx, `{"t":"d","d":{"r":1,"a":"l","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d","h":""}}}`)
	send(c, appCtx, `{"t":"d","d":{"r":2,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44","d":{"fullname":"go-cli","id":"fe24478e-0161-0c97-18ef-ab569207ac44"}}}}`)
	send(c, appCtx, `{"t":"d","d":{"r":3,"a":"o","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":{".sv":"timestamp"}}}}`)
	send(c, appCtx, `{"t":"d","d":{"r":4,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":true}}}`)
	//if err != nil {
	//	msg := fmt.Sprintf("Unable to write to websocket: %v", err)
	//	fmt.Println(msg)
	//}
}

func connect(appCtx context.Context) *websocket.Conn {
	ctx, cancel := context.WithTimeout(appCtx, 5*time.Second)
	defer cancel()

	fmt.Println("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to websocket: %v", err)
		fmt.Println(msg)
		panic(err)
	}
	return c
}

func keepalive(c *websocket.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := send(c, ctx, "0")
	if err != nil {
		msg := fmt.Sprintf(">>> Unable to send keepalive: %v", err)
		fmt.Println(msg)
	}
}

func send(c *websocket.Conn, ctx context.Context, payload string) error {
	println(">>> " + payload)
	return c.Write(ctx, websocket.MessageText, []byte(payload))
}
