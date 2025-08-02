package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fmt.Println("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to websocket: %v", err)
		fmt.Println(msg)
		panic(err)
	}
	defer c.CloseNow()

	fmt.Println("Sending keepalive ('0')")
	err = wsjson.Write(ctx, c, "0")
	if err != nil {
		msg := fmt.Sprintf("Unable to write to websocket: %v", err)
		fmt.Println(msg)
	}

	fmt.Println("Joining game (1)")
	c.Write(ctx, websocket.MessageText, []byte(`{"t":"d","d":{"r":1,"a":"l","b":{"p":"/games/b2bb45fb-1fa8-889c-99d8-f1f996b9c568","h":""}}}`))
	c.Write(ctx, websocket.MessageText, []byte(`{"t":"d","d":{"r":2,"a":"p","b":{"p":"/games/b2bb45fb-1fa8-889c-99d8-f1f996b9c568/participants/fe24478e-0161-0c97-18ef-ab569207ac44","d":{"fullname":"go-cli","id":"fe24478e-0161-0c97-18ef-ab569207ac44"}}}}`))
	c.Write(ctx, websocket.MessageText, []byte(`{"t":"d","d":{"r":3,"a":"o","b":{"p":"/games/b2bb45fb-1fa8-889c-99d8-f1f996b9c568/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":{".sv":"timestamp"}}}}`))
	c.Write(ctx, websocket.MessageText, []byte(`{"t":"d","d":{"r":4,"a":"p","b":{"p":"/games/b2bb45fb-1fa8-889c-99d8-f1f996b9c568/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":true}}}`))
	//if err != nil {
	//	msg := fmt.Sprintf("Unable to write to websocket: %v", err)
	//	fmt.Println(msg)
	//}

	fmt.Println("Reading response (1)")
	mt, bytes, err := c.Read(ctx)
	if err != nil {
		msg := fmt.Sprintf("Unable to read from websocket: %v", err)
		fmt.Println(msg)
	} else {
		println("MessageType is ", mt.String())
		println(string(bytes))
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Hit Enter to exit program: ")
	scanner.Scan()

	fmt.Println("Closing websocket (using handshake)")
	c.Close(websocket.StatusNormalClosure, "")
}
