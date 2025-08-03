package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/coder/websocket"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	//ctx30, cancel30 := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel30()
	//
	fmt.Println("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to websocket: %v", err)
		fmt.Println(msg)
		panic(err)
	}
	defer c.CloseNow()

	go func() {
		//ctx30, cancel30 := context.WithTimeout(context.Background(), 30*time.Second)
		//defer cancel30()
		for {
			//fmt.Println("Reading response(s) (1)")
			_, bytes, err := c.Read(context.Background())
			if err != nil {
				msg := fmt.Sprintf("<<< Unable to read from websocket: %v", err)
				fmt.Println(msg)
				break
			} else {
				//println("MessageType is ", mt.String())
				println("<<< " + string(bytes))
			}
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			keepalive(c)
		}
	}()

	//fmt.Println("Joining game (1)")
	send(c, ctx, `{"t":"d","d":{"r":1,"a":"l","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d","h":""}}}`)
	send(c, ctx, `{"t":"d","d":{"r":2,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44","d":{"fullname":"go-cli","id":"fe24478e-0161-0c97-18ef-ab569207ac44"}}}}`)
	send(c, ctx, `{"t":"d","d":{"r":3,"a":"o","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":{".sv":"timestamp"}}}}`)
	send(c, ctx, `{"t":"d","d":{"r":4,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":true}}}`)
	//if err != nil {
	//	msg := fmt.Sprintf("Unable to write to websocket: %v", err)
	//	fmt.Println(msg)
	//}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Hit Enter to exit program: ")
	scanner.Scan()

	fmt.Println("Closing websocket (using handshake)")
	c.Close(websocket.StatusNormalClosure, "")
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
