/*
 * // Copyright (c) 2020 @thiinbit. All rights reserved.
 * // Use of this source code is governed by an MIT-style
 * // license that can be found in the LICENSE file
 *
 */

package gosocket

import (
	"log"
	"testing"
	"time"
)

func TestTCPClient_SendMessage(t *testing.T) {
	server, _ := NewTCPServer("0.0.0.0:8888").
		//server, _ := NewTCPServer("[::1]:8888").
		RegisterMessageListener(&BroadcastServerMessageListener{}). // Required
		Run()

	go func() {
		<-time.NewTimer(20 * time.Second).C

		_ = server.Stop()
	}()

	client, _ := NewTcpClient("[::1]:8888").
		RegisterMessageListener(&TestExampleClientListener{}).
		Dial()

	_ = client.SendMessage("...Hello!")

	client.Hangup("Test sendMessage wrong status!")

	err := client.SendMessage("...After hangup message!")
	if err != nil {
		log.Print("......", err)
	}

	client, _ = client.Dial()
	err2 := client.SendMessage("...Redialed!")
	log.Print("...Send Redialed")
	if err2 != nil {
		log.Print("......", err2)
	}

	client.Hangup("...Test")
	_ = server.Stop()
}
