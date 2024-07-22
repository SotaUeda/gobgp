package main

import (
	"fmt"
	"os"
	"time"

	"github.com/SotaUeda/gobgp/peer"
)

func main() {
	confStrs := []string{
		"64512 127.0.0.1 65413 127.0.0.2 active",
	}
	var peers []peer.Peer
	for _, s := range confStrs {
		c, err := peer.ParseConfig(s)
		if err != nil {
			fmt.Printf("Config Error: %v\n", err)
			os.Exit(1)
		}
		peers = append(peers, *peer.NewPeer(c))
	}
	for _, p := range peers {
		p.Start()
	}

	for _, p := range peers {
		for {
			p.Next()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Context必要？
