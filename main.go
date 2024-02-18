package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/olahol/melody"

	"socket/models"
)

// type GopherInfo struct {
// 	ID, X, Y string
// }

func main() {
	m := melody.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		// ss, _ := m.Sessions()

		// for _, o := range ss {
		// value, exists := o.Get("info")

		// if !exists {
		// 	continue
		// }

		// info := value.(*GopherInfo)

		// s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
		// }

		id := uuid.NewString()
		// s.Set("info", &GopherInfo{id, "0", "0"})

		s.Write([]byte("iam " + id))
		log.Println(id + " connected")
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
		if err := message.Decode(string(msg)); err != nil {
			log.Panicln(err)
			return
		}

		// fmt.Println(message)

		// find channel
		channel := models.Channel{}
		if err := channel.FirstOrCreate(message.Channel); err != nil {
			log.Panicln(err)
			return
		}

		// handle action of message
		if message.Action == "join" {

			channel.Join(message.Username, message.Channel)

		} else if message.Action == "leave" {

			channel.Leave(message.Username, message.Channel)

		}

		// json encode message
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		// get channel info
		channelInfo := channel.InfoMessage()

		// message to sender about channel info
		s.Write([]byte(channelInfo))

		// return message to sender
		s.Write([]byte(response))

		m.BroadcastOthers([]byte(response), s)
	})

	http.ListenAndServe(":8000", nil)
}
