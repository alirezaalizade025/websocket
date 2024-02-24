package models

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/olahol/melody"
)

type Channel struct {
	ID          int    `json:"id"`
	ChannelName string `json:"channel_name"`
	// ChannelClients []ChannelClient `json:"users"`
}

type Status string

// const (
// 	Active   Status = "active"
// 	Inactive Status = "inactive"
// )

// type ChannelClient struct {
// 	ID     string `json:"id"`
// 	Status Status `json:"status"`
// }

// var Channels = map[string]*Channel{}
var Channels sync.Map

func (c *Channel) Join(id string) {

	// if exists skip
	for _, item := range ChannelClients {
		if item.ClientID == id && item.ChannelID == c.ID {
			return
		}
	}

	ChannelClients = append(ChannelClients, &ChannelClient{
		ClientID:  id,
		ChannelID: c.ID,
		Status:    Active,
		JoinAt:    time.Now().Format(time.RFC3339),
	})

	// for _, client := range c.ChannelClients {
	// 	if client.ID == id {
	// 		return
	// 	}
	// }

	// _, found := ChannelClients.Get("id")
	// if found {
	// 	return
	// }

	// c.ChannelClients = append(c.ChannelClients, ChannelClient{
	// 	ID:     id,
	// 	Status: Active,
	// })

	// Channels[c.ChannelName] = &Channel{
	// 	ChannelName:    c.ChannelName,
	// 	ChannelClients: c.ChannelClients,
	// }

	// // add channel to client channels
	// for i, item := range Clients {
	// 	if item.ID == id {
	// 		Clients[i].Channels = append(Clients[i].Channels, c.ChannelName)
	// 	}
	// }
}

func (c *Channel) Leave(id string) {

	if id == "" {
		return
	}

	for i, item := range ChannelClients {
		if item.ClientID == id && item.ChannelID == c.ID {
			ChannelClients = append(ChannelClients[:i], ChannelClients[i+1:]...)
		}
	}

	// go func() {
	var channelClientsCount int
	for _, item := range ChannelClients {
		if item.ChannelID == c.ID {
			channelClientsCount++
		}
	}

	if channelClientsCount == 0 {
		c.RemoveChannel()
	}
	// }()

	// for i, u := range c.ChannelClients {
	// 	if u.ID == id {
	// 		c.ChannelClients = append(c.ChannelClients[:i], c.ChannelClients[i+1:]...)
	// 	}
	// }

	// Channels[c.ChannelName] = &Channel{
	// 	ChannelName:    c.ChannelName,
	// 	ChannelClients: c.ChannelClients,
	// }

	// // remove channel name from client channels
	// for i, item := range Clients {
	// 	if item.ID == id {
	// 		for j, ch := range Clients[i].Channels {
	// 			if ch == c.ChannelName {
	// 				Clients[i].Channels = append(Clients[i].Channels[:j], Clients[i].Channels[j+1:]...)
	// 			}
	// 		}
	// 	}
	// }

	// if len(c.ChannelClients) == 0 {
	// 	c.RemoveChannel()
	// }
}

func (c *Channel) RemoveChannel() {

	// remove channel from channels
	// delete(Channels, c.ChannelName)

	Channels.Delete(c.ChannelName)

}

func (c *Channel) ActiveClient(id string) {

	if id == "" {
		return
	}

	for i, item := range ChannelClients {
		if item.ClientID == id && item.ChannelID == c.ID {
			ChannelClients[i].Status = Active
		}
	}

	// for i, ChannelClient := range c.ChannelClients {
	// 	if ChannelClient.ID == id {
	// 		c.ChannelClients[i].Status = Active
	// 	}
	// }

	// Channels[c.ChannelName] = &Channel{
	// 	ChannelName:    c.ChannelName,
	// 	ChannelClients: c.ChannelClients,
	// }

}

func (c *Channel) InactiveClient(id string) {

	if id == "" {
		return
	}

	for i, item := range ChannelClients {
		if item.ClientID == id && item.ChannelID == c.ID {
			ChannelClients[i].Status = Inactive
		}
	}

	// for i, ChannelClient := range c.ChannelClients {
	// 	if ChannelClient.ID == id {
	// 		c.ChannelClients[i].Status = Inactive
	// 	}
	// }

	// Channels[c.ChannelName] = &Channel{
	// 	ChannelName:    c.ChannelName,
	// 	ChannelClients: c.ChannelClients,
	// }
}

func (c *Channel) FirstOrCreate(channelName string) error {

	if !channelExists(channelName) {

		newChannel := Channel{
			ID:          autoIncrementId(),
			ChannelName: channelName,
		}

		// Channels[channelName] = &newChannel

		Channels.Store(channelName, &newChannel)

	}

	channel, _ := Channels.Load(channelName)

	c.ID = channel.(*Channel).ID
	c.ChannelName = channelName

	if c.ChannelName == "" {
		return errors.New("channel not found")
	}

	return nil
}

// function autoIncrementId
// get max id of Channels and return +1
func autoIncrementId() int {
	max := 0

	Channels.Range(func(key, value interface{}) bool {
		if value.(*Channel).ID > max {
			max = value.(*Channel).ID
		}
		return true
	})

	return max + 1
}

func (c *Channel) InfoMessage() []byte {

	var joinedUsernames []map[string]interface{}

	// for _, ChannelClient := range c.ChannelClients {
	// 	client := FindByID(ChannelClient.ID)

	// 	if client.Username == "" {
	// 		// 	user.Username = id
	// 		panic("username not found")
	// 	}

	// 	joinedUsernames = append(joinedUsernames, map[string]interface{}{
	// 		"username":   client.Username,
	// 		"connect_at": client.ConnectAt,
	// 		"avatar":     client.Avatar,
	// 		"status":     ChannelClient.Status,
	// 	})
	// }

	for i, item := range ChannelClients {
		if item.ChannelID == c.ID {
			client := FindByID(item.ClientID)

			if client.Username == "" {
				// 	user.Username = id
				ChannelClients = append(ChannelClients[:i], ChannelClients[i+1:]...)

				log.Panic("username not found")
			}

			joinedUsernames = append(joinedUsernames, map[string]interface{}{
				"username":   client.Username,
				"connect_at": client.ConnectAt,
				"avatar":     client.Avatar,
				"status":     item.Status,
			})
		}
	}

	type Data struct {
		// ChannelName    string                   `json:"channel_name"`
		ChannelClients []map[string]interface{} `json:"users"`
	}

	info, err := json.Marshal(Message{
		ChannelName: c.ChannelName,
		Action:      "channel_info",
		Data: Data{
			// ChannelName:    c.ChannelName,
			ChannelClients: joinedUsernames,
		},
	})

	if err != nil {
		log.Panicln(err)
	}
	return info
}

func (c *Channel) InChannel(id string) bool {
	// for _, u := range c.ChannelClients {
	// 	if u.ID == id && u.Status == Active {
	// 		return true
	// 	}
	// }
	// return false

	for _, item := range ChannelClients {
		if item.ClientID == id && item.ChannelID == c.ID && item.Status == Active {
			return true
		}
	}

	return false
}

func (c *Channel) Broadcast(message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return c.InChannel(q.Keys["id"].(string)) || q.Keys["id"].(string) == AdminClient.ID
	})
}

func (c *Channel) BroadcastOther(s *melody.Session, message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return (c.InChannel(q.Keys["id"].(string)) && q != s) || q.Keys["id"].(string) == AdminClient.ID
	})
}

func ChannelFirst(channelName string) (c *Channel, err error) {

	if channelExists(channelName) {

		c, found := Channels.Load(channelName)
		if !found {
			return c.(*Channel), errors.New("channel not found")
		}

		return c.(*Channel), nil
	}

	return c, errors.New("channel not found")
}

func ChannelFirstById(id int) (c Channel, err error) {

	Channels.Range(func(key, value interface{}) bool {
		if value.(*Channel).ID == id {
			c = *value.(*Channel)
			return false
		}
		return true
	})

	if c.ChannelName != "" {
		return c, nil
	}

	return c, errors.New("channel not found")
}

func channelExists(name string) bool {

	_, found := Channels.Load(name)
	return found

}
