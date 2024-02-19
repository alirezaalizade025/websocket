package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/olahol/melody"

	"socket/models"
)

func main() {
	m := melody.New()

	m.Config.PingPeriod = 10 * time.Second

	// log.Println(m.Config.PingPeriod)


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

		client.LeaveAllChannels(m)
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
		

		if message.Action == "join" || message.Action == "leave" {
			channelInfo := channel.InfoMessage()
			channel.Broadcast(channelInfo, m)
			return
		} 

		// json encode message
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		// // return message to sender
		// s.Write([]byte(response))

		channel.BroadcastOther(s, response, m)
	})

	m.HandlePong(func(s *melody.Session) {

		s.Write([]byte("Ping"))

		

		log.Println("Pong received", s.IsClosed(), s.Keys["id"])
	})


	log.Println("Server started at :8000")
	http.ListenAndServe(":8000", nil)
}
