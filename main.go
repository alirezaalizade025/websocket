package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/olahol/melody"

	"socket/models"
)

func main() {
	m := melody.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		client := models.NewClient()

		s.Keys = map[string]interface{}{"id": client.ID}

	})

	m.HandleDisconnect(func(s *melody.Session) {

		// if key is empty return
		if s.Keys["id"] == nil {
			return
		}

		client := models.FindByID(s.Keys["id"].(string))

		clientChannels := client.Channels
		for _, channelName := range clientChannels {
			channel, err := models.ChannelFirst(channelName)
			if err != nil {
				continue
			}

			channel.Leave(s.Keys["id"].(string))

			// // ---- leave broadcast -----
			// message, err := json.Marshal(models.Message{
			// 	Username:    client.Username,
			// 	ChannelName: channelName,
			// 	Action:      "disconnect",
			// })
			// if err != nil {
			// 	log.Panicln(err)
			// }
			// channel.Broadcast(message, m)

			// ---- channel info broadcast -----
			channel.Broadcast(channel.InfoMessage(), m)
		}

	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {

		// decode message
		message := models.Message{}
		if err := message.Decode(msg); err != nil {
			log.Panicln(err)
			return
		}

		// find channel
		channel := models.Channel{}
		err := channel.FirstOrCreate(message.ChannelName)
		if err != nil {
			log.Panicln(err)
			return
		}

		models.MatchUsernameWithID(s.Keys["id"].(string), message.Username)

		// handle action of message
		if message.Action == "join" {

			channel.Join(s.Keys["id"].(string))

			// match username with id
		} else if message.Action == "leave" {

			channel.Leave(s.Keys["id"].(string))

		}

		// log.Println(models.Clients)

		// get channel info
		channelInfo := channel.InfoMessage()
		// message to sender about channel info
		// s.Write([]byte(channelInfo))
		// m.Broadcast([]byte(channelInfo))
		channel.Broadcast(channelInfo, m)

		if message.Action == "join" || message.Action == "leave" {
			return
		}

		// json encode message
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		// return message to sender
		s.Write([]byte(response))

		// m.BroadcastOthers([]byte(response), s)

		channel.Broadcast(response, m)
	})

	// m.HandlePong(func(s *melody.Session) {
	// 	log.Println("Pong received")
	// })

	log.Println("Server started at :8000")
	http.ListenAndServe(":8000", nil)
}
