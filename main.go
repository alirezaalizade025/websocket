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
	"socket/utils"
)

const (
	PingPeriod  = 8 * time.Second
	PongTimeOut = 8 * PingPeriod / 10
)

var ws = melody.New()

func main() {

	// utils.LoadDotEnv()

	ws.Config.MaxMessageSize = 1024 * 10

	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		ws.HandleRequest(c.Writer, c.Request)
	})

	r.Use(middlewares.Cors)

	r.POST("api/broadcast", middlewares.BasicAuth, func(c *gin.Context) {
		controllers.Broadcast(c, ws)
	})

	ws.HandleConnect(func(s *melody.Session) {

		client := models.NewClient()

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

		ws.Broadcast([]byte(models.ClientsInfo()))

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
			pong(s, message)

			return
		}

		if message.Action == "init" {

			initClient(s, message)

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
			s.Set("status", models.Active)
			client.SetStatus(models.Active)
			return
		case "inactive":
			client := models.FindByID(s.Keys["id"].(string))
			client.InactiveAllChannels(ws, s)
			s.Set("status", models.Inactive)
			client.SetStatus(models.Inactive)
			return
		}

		// peer to peer
		if message.Action == "p2p" {
			p2p(message)
			return
		}

		// find channel
		channel := models.Channel{}
		err := channel.FirstOrCreate(*message.ChannelName)
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

	ws.HandleError(func(s *melody.Session, err error) {
		log.Println("Session error", err)
	})

	go unResponsesPong()

	r.Run(":8000")
}

func p2p(message models.Message) {
	if message.Data == nil {
		log.Println("p2p message data is nil")
	}

	if message.Username == nil {
		log.Println("p2p message username is nil")
	}

	clients := models.FindByUsername(*message.Username)

	if len(clients) == 0 {
		log.Println("p2p message username not found")
	}

	receiversIds := []string{}
	var activeClientIds []string
	var inactiveClientIds []string

	for _, client := range clients {

		if client.ID != "" {

			if client.Status == models.Active {
				activeClientIds = append(activeClientIds, client.ID)
			} else {
				inactiveClientIds = append(inactiveClientIds, client.ID)
			}
		}

		clientReceiversIds := handleSendMode("active_first", activeClientIds, inactiveClientIds)
		receiversIds = append(receiversIds, clientReceiversIds...)
		
	}

	response, err := json.Marshal(message)
	if err != nil {
		log.Panicln(err)
	}

	ws.BroadcastFilter(response, func(s *melody.Session) bool {
		return utils.Contains(receiversIds, s.Keys["id"].(string))
	})
}

func handleSendMode(mode string, activeClientIds, inactiveClientIds []string) (receiversIds []string) {
	if mode == "active_only" {
		receiversIds = activeClientIds
	} else if mode == "active_first" {
		if len(activeClientIds) > 0 {
			receiversIds = activeClientIds
		} else {
			receiversIds = inactiveClientIds
		}
	} else {
		receiversIds = append(activeClientIds, inactiveClientIds...)
	}
	return receiversIds
}

func initClient(s *melody.Session, message models.Message) {
	client := models.FindByID(s.Keys["id"].(string))
	client.UpdateClient(message)

	s.Write([]byte(client.InitMessage()))

	ws.Broadcast([]byte(models.ClientsInfo()))
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

			if client.(*models.Client).Username == "" {
				s.Close()
				continue
			}

			time := time.Now()
			client.(*models.Client).PingAt = &time
		}
	}()
}

func pong(s *melody.Session, message models.Message) {

	client, found := models.Clients.Load(s.Keys["id"].(string))
	if !found {
		client = models.NewClient()
	}

	client.(*models.Client).PingAt = nil

	if message.Data == nil || message.Data.(map[string]interface{})["status"] == nil {
		return
	}

	user := client.(*models.Client)

	if message.Data.(map[string]interface{})["status"].(string) == "active" && user.Status == models.Inactive {

		user.ActiveAllChannels(ws, s)
		s.Set("status", models.Active)
		user.SetStatus(models.Active)

	} else if message.Data.(map[string]interface{})["status"].(string) == "inactive" && user.Status == models.Active {

		user.InactiveAllChannels(ws, s)
		s.Set("status", models.Inactive)
		user.SetStatus(models.Inactive)
	}

	// log.Println("Pong received", s.IsClosed(), s.Keys["id"])

}

func unResponsesPong() {

	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {

		models.Clients.Range(func(key, value interface{}) bool {
			client := value.(*models.Client)

			if client.ID != models.AdminClient.ID &&
				client.PingAt != nil &&
				time.Since(client.ConnectAt) > PingPeriod &&
				time.Since(*client.PingAt) > PongTimeOut {

				session, err := client.GetSession(ws)

				if err != nil {

					if err.Error() == "session not found" {
						client.Delete()
						return true
					}

					log.Println("Get Session: " + err.Error())
					return true
				}

				client.InactiveAllChannels(ws, session)

				if !session.IsClosed() {

					err = session.Close()
					if err != nil {
						log.Println("Close Session: " + err.Error())
					}
				}
			}

			return true
		})

	}

}
