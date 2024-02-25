package main

import (
	"encoding/json"
	"log"
	"time"

	// "time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"socket/controllers"
	"socket/middlewares"
	"socket/models"
	// "socket/utils"
)

const (
	PingPeriod  = 10 * time.Second
	PongTimeOut = 8 * PingPeriod / 10
)

var ws = melody.New()

func main() {

	// utils.LoadDotEnv()

	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		ws.HandleRequest(c.Writer, c.Request)
	})

	r.POST("/broadcast", middlewares.BasicAuth, func(c *gin.Context) {
		controllers.Broadcast(c, ws)
	})

	ws.HandleConnect(func(s *melody.Session) {

		client := models.NewClient()

		// s.Keys = map[string]interface{}{"id": client.ID}
		s.Set("id", client.ID)

		go ping(s)
	})

	ws.HandleDisconnect(func(s *melody.Session) {

		// if key is empty return
		if s.Keys["id"] == nil {
			return
		}

		client := models.FindByID(s.Keys["id"].(string))

		client.LeaveAllChannels(ws)

		client.Delete()

		s.Close()
	})

	ws.HandleMessage(func(s *melody.Session, msg []byte) {

		// decode message
		message := models.Message{}
		if err := message.Decode(msg); err != nil {
			log.Println(err)
			return
		}

		if message.Action == "pong" {
			pong(s)
			return
		}

		if message.Action == "init" {
			client := models.FindByID(s.Keys["id"].(string))
			client.UpdateClient(message)

			s.Write([]byte(client.InitMessage()))

			return
		}

		if message.Action == "init_admin" {
			client := models.FindByID(s.Keys["id"].(string))

			client.InitAdmin()

			s.Write([]byte(client.AdminInitMessage()))

			models.Channels.Range(func(key, value interface{}) bool {
				channel := value.(*models.Channel)
				s.Write([]byte(channel.InfoMessage()))
				return true
			})

			return
		}

		switch message.Action {
		case "active":
			client := models.FindByID(s.Keys["id"].(string))
			client.ActiveAllChannels(ws, s)
			return
		case "inactive":
			client := models.FindByID(s.Keys["id"].(string))
			client.InactiveAllChannels(ws, s)
			return
		}

		// find channel
		channel := models.Channel{}
		err := channel.FirstOrCreate(message.ChannelName)
		if err != nil {
			log.Panicln(err)
			return
		}

		// handle action of message
		switch message.Action {
		case "join":
			channel.Join(s.Keys["id"].(string))

			channelInfo := channel.InfoMessage()
			channel.Broadcast(channelInfo, ws)
			return

		case "leave":
			channel.Leave(s.Keys["id"].(string)) // todo: delete from clients

			channelInfo := channel.InfoMessage()
			channel.Broadcast(channelInfo, ws)
			return
		}

		// json encode message
		response, err := json.Marshal(message)
		if err != nil {
			log.Panicln(err)
		}

		channel.BroadcastOther(s, response, ws)
	})

	// m.HandleClose(func(s1 *melody.Session, i int, s2 string) error {

	// 	log.Println("Session closed", s1.IsClosed(), s1.Keys["id"])
	// 	return nil
	// })

	ws.HandleError(func(s *melody.Session, err error) {
		log.Println("Session error", err)
	})

	// m.HandleSentMessage(func(s *melody.Session, msg []byte) {

	// 	log.Println("Sent message", string(msg))
	// })

	// m.HandlePong(func(s *melody.Session) {

	// 	log.Println("Pong received", s.IsClosed(), s.Keys["id"])

	// })

	go unResponsesPong()

	r.Run(":8000")
}

func ping(s *melody.Session) {

	ticker := time.NewTicker(PingPeriod)

	ping, err := json.Marshal(map[string]string{
		"action": "ping",
	})

	if err != nil {
		log.Panicln(err)
	}

	go func() {
		for range ticker.C {
			s.Write(ping)

			client, found := models.Clients.Load(s.Keys["id"].(string))
			if !found {
				break
			}

			time := time.Now()
			client.(*models.Client).PingAt = &time
		}
	}()
}

func pong(s *melody.Session) {

	client, found := models.Clients.Load(s.Keys["id"].(string))
	if !found {
		log.Println("new")
		client = models.NewClient()
	}

	client.(*models.Client).PingAt = nil

	// log.Println("Pong received", s.IsClosed(), s.Keys["id"])

}

func unResponsesPong() {

	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {

		models.Clients.Range(func(key, value interface{}) bool {
			client := value.(*models.Client)


			if client.PingAt != nil &&
				time.Since(client.ConnectAt) > PingPeriod &&
				time.Since(*client.PingAt) > PongTimeOut {

				session, err := client.GetSession(ws)

				if err != nil {

					if err.Error() == "session not found" {
						client.Delete()
						return true
					}

					log.Println(err)
					return true
				}

				client.InactiveAllChannels(ws, session)


				err = session.Close()
				if err != nil {
					log.Println(err)
				}
			}

			return true
		})

	}

}
