// Command chat runs the chat server.
//
//	go run ./phase-8-portfolio/day-087-chat-server/starter/cmd/chat -addr :8080
package main

import (
	"fmt"

	chat "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-087-chat-server/starter"
)

func main() {
	// Your code here: parse -addr, start a hub with hub.Run in a goroutine, serve
	// chat.NewServer(hub) at /ws, and shut down cleanly on Ctrl+C.
	fmt.Println(chat.SendBuffer)
}
