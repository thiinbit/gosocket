// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package main

import (
	"bufio"
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
				&cli.StringFlag{
					Name:    "debug",
					Aliases: []string{"d"},
					Usage:   "open debug log, true|false",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("listen") == "" {
					_ = cli.ShowCommandHelp(c, "server")
					return nil
				}

				debug := false
				if c.String("debug") == "true" {
					debug = true
				}

				server, err := gosocket.NewTCPServer(c.String("listen")).
					RegisterMessageListener(&gosocket.BroadcastServerMessageListener{}).
					SetDebugMode(debug).
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
				&cli.StringFlag{
					Name:    "debug",
					Aliases: []string{"d"},
					Usage:   "open debug log, true|false",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("target") == "" {
					_ = cli.ShowCommandHelp(c, "client")
					return nil
				}

				debug := false
				if c.String("debug") == "true" {
					debug = true
				}

				client, err := gosocket.NewTcpClient(c.String("target")).
					RegisterMessageListener(&gosocket.ExampleClientMessageListener{}).
					SetDebugMode(debug).
					Dial()
				if err != nil {
					return err
				}

				reader := bufio.NewReader(os.Stdin)
				for {
					text, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					if text == "quit()\n" {
						break
					}
					err = client.SendMessage(text)
					if err != nil {
						return err
					}
				}

				client.Hangup("Quit")
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
