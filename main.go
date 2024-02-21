package main

import (
	"encoding/json"
	"log"

	// "time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"socket/controllers"
	"socket/middlewares"
	"socket/models"
	"socket/utils"
)

func main() {

	utils.LoadDotEnv()

	r := gin.New()
	m := melody.New()

	// m.Config.PingPeriod = 9 * time.Second
	// m.Config.PongWait = 10 * time.Second

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	r.POST("/broadcast", middlewares.BasicAuth, func(c *gin.Context) {
		controllers.Broadcast(c, m)
	}, )

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

		// log.Println("Session disconnected", s.IsClosed(), s.Keys["id"])
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {

		// decode message
		message := models.Message{}
		if err := message.Decode(msg); err != nil {
			log.Panicln(err)
			return
		}

		if message.Action == "init" {
			client := models.FindByID(s.Keys["id"].(string))
			client.UpdateClient(message)

			s.Write([]byte(client.InitMessage()))

			return
		}

		switch message.Action {
		case "active":
			client := models.FindByID(s.Keys["id"].(string))
			client.ActiveAllChannels(m, s)
			return
		case "inactive":
			client := models.FindByID(s.Keys["id"].(string))
			client.InactiveAllChannels(m, s)
			return
		}

		// find channel
		channel := models.Channel{}
		err := channel.FirstOrCreate(message.ChannelName)
		if err != nil {
			log.Panicln(err)
			return
		}

		// models.MatchUsernameWithID(s.Keys["id"].(string), message.Username)

		// handle action of message
		switch message.Action {
		case "join":
			channel.Join(s.Keys["id"].(string))

			channelInfo := channel.InfoMessage()
			channel.Broadcast(channelInfo, m)
			return

		case "leave":
			channel.Leave(s.Keys["id"].(string))

			channelInfo := channel.InfoMessage()
			channel.Broadcast(channelInfo, m)
			return
		}

		// json encode message
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		// return message to sender
		// s.Write([]byte(response))

		channel.BroadcastOther(s, response, m)
	})

	m.HandlePong(func(s *melody.Session) {

		log.Println("Pong received", s.IsClosed(), s.Keys["id"])

	})

	m.HandleClose(func(s1 *melody.Session, i int, s2 string) error {

		// log.Println("Session closed", s1.IsClosed(), s1.Keys["id"])
		return nil
	})

	m.HandleError(func(s *melody.Session, err error) {
		log.Println("Session error", err)
	})

	// m.HandleSentMessage(func(s *melody.Session, msg []byte) {

	// 	log.Println("Sent message", string(msg))
	// })

	// ticker := time.NewTicker(m.Config.PingPeriod)

	// go func() {
	// 	for range ticker.C {
	// 		m.Broadcast([]byte("ping"))
	// 	}
	// }()

	// m.HandlePong(func(s *melody.Session) {
	// 	log.Println("Pong received", s.IsClosed(), s.Keys["id"])

	// })

	r.Run(":8000")
}
