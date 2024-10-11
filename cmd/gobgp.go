package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SotaUeda/gobgp/peer"
)

func main() {
	// 引数で与えられた文字列を順に結合してconfig文字列を作成
	config := os.Args[1]
	confStrs := []string{
		config,
	}
	var peers []peer.Peer
	for _, s := range confStrs {
		c, err := peer.ParseConfig(s)
		if err != nil {
			fmt.Printf("Config Error: %v\n", err)
			os.Exit(1)
		}
		// LocRibはすべてのPeerで共有する
		// 排他制御のためにsync.Mutexを使う
		locRib, err := peer.NewLocRib(c)
		if err != nil {
			fmt.Printf("LocRib Error: %v\n", err)
			os.Exit(1)
		}
		var mu sync.Mutex
		mu.Lock()
		peers = append(peers, *peer.NewPeer(c, locRib))
		mu.Unlock()
	}
	for _, p := range peers {
		p.Start()
	}

	ctx, cansel := context.WithCancel(context.Background())
	for _, p := range peers {
		go func() {
			for {
				if err := p.Next(ctx); err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	// Ctrl-c入力時にプログラムを停止
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cansel()
	}()
	<-ctx.Done()
}
