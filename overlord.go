package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/giongto35/cloud-game/webrtc"
)

var overlordRooms = map[string]string{}

// servers are the map serverID to server Client
var servers = map[string]Client{}

// If it's overlord, handle overlord connection (from host to overlord)
func wso(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connected")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("[!] WS upgrade:", err)
		return
	}
	defer c.Close()

	// register new server
	serverID := strconv.Itoa(rand.Int())
	log.Println("A new server connected ", serverID)

	client := NewClient(c, webrtc.NewWebRTC())

	client.send(
		WSPacket{
			ID:   "serverID",
			Data: serverID,
		},
	)

	client.syncReceive("ping", func(resp WSPacket) WSPacket {
		log.Println("received Ping, sending Pong")
		return WSPacket{
			ID: "pong",
		}
	})

	client.syncReceive("registerRoom", func(resp WSPacket) WSPacket {
		log.Println("received registerRoom")
		overlordRooms[resp.Data] = serverID
		return WSPacket{
			ID: "registerRoom",
		}
	})

	client.listen()
}