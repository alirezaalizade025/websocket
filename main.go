package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/olahol/melody"

	"socket/models"
)


func main() {
	m := melody.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		// ss, _ := m.Sessions()

		id := uuid.NewString()

		// s.Write([]byte("iam " + s.Request.URL.Query().Get("username")))
		log.Println(id + " connected")

		s.Keys = make(map[string]interface{}) // Initialize the Keys map

		s.Keys["id"] = id
	})

	m.HandleDisconnect(func(s *melody.Session) {
		// value, exists := s.Get("info")

		// if !exists {
		// 	return
		// }

		// info := value.(*GopherInfo)

		// m.BroadcastOthers([]byte("dis "+info.ID), s)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {

		// decode message
		message := models.Message{}
		if err := message.Decode(msg); err != nil {
			log.Panicln(err)
			return
		}

		log.Println(s.Keys)
		log.Println(message)

		// find channel
		channel := models.Channel{}
		err := channel.FirstOrCreate(message.Channel)
		if err != nil {
			log.Panicln(err)
			return
		}

		// handle action of message
		if message.Action == "join" {
			channel.Join(message.Username, message.Channel)
		} else if message.Action == "leave" {
			channel.Leave(message.Username, message.Channel)
		}

		// get channel info
		channelInfo := channel.InfoMessage()
		// message to sender about channel info
		s.Write([]byte(channelInfo))
		m.Broadcast([]byte(channelInfo))


		// json encode message 
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		// return message to sender
		s.Write([]byte(response))

		m.BroadcastOthers([]byte(response), s)
	})

	log.Println("Server started at :8000")
	http.ListenAndServe(":8000", nil)
}
