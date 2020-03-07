// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package main

import (
	"context"
	"github.com/thiinbit/gosocket"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	app := cli.NewApp()
	app.Name = "Gosocket"
	app.Version = "0.0.1"
	app.Usage = "A simple, session, heartbeat socket library."
	app.Authors = []*cli.Author{
		{
			Name:  "@thiinbit",
			Email: "thiinbit@gmail.com",
		},
	}
	app.Copyright = "https://github.com/thiinbit/gosocket"
	app.Commands = []*cli.Command{
		{
			Name:  "server",
			Usage: "Run as a broadcast server",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "listen",
					Aliases: []string{"l"},
					Usage:   "Server listen address, like: 0.0.0.0:8888",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("listen") == "" {
					_ = cli.ShowCommandHelp(c, "server")
					return nil
				}
				server, err := gosocket.NewTCPServer(c.String("listen")).
					RegisterMessageListener(&BroadcastServerMessageListener{}).
					Run()
				if err != nil {
					return err
				}

				sig := make(chan os.Signal, 1)
				go func() {
					signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
				}()
				<-sig

				return server.Stop()
			},
		},
		{
			Name:  "client",
			Usage: "Run as a client",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "target",
					Aliases: []string{"t"},
					Usage:   "Dial up to target server, like: 127.0.0.1:8888",
				},
			},
			Action: func(c *cli.Context) error {
				print("Not support now!")
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// BroadcastServerMessageListener
type BroadcastServerMessageListener struct{}

// BroadcastServerMessageListener impl
func (tl *BroadcastServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *gosocket.Session) {
	log.Print("Server received message: ", message)

	for k, v := range session.ServerRef().Sessions() {
		if session.SID() != k {
			log.Printf("Broadcast message to client %s: %s ", v.SID(), message)
			v.SendMessage(message)
		}
	}
}
