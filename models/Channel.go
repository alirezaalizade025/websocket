package models

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/olahol/melody"
)

type Channel struct {
	ChannelName  string   `json:"channel_name"`
	ChannelUsers []string `json:"users"`
}

var Channels = map[string]*Channel{}

func (c *Channel) Join(id string) {

	// append if not exists
	for _, u := range c.ChannelUsers {
		if u == id {
			return
		}
	}

	c.ChannelUsers = append(c.ChannelUsers, id)


	Channels[c.ChannelName] = &Channel{
		ChannelName:  c.ChannelName,
		ChannelUsers: c.ChannelUsers,
	}

	// add channel to client channels
	for i, item := range Clients {
		if item.ID == id {
			Clients[i].Channels = append(Clients[i].Channels, c.ChannelName)
		}
	}
}

func (c *Channel) Leave(id string) {

	if id == "" {
		return
	}

	for i, u := range c.ChannelUsers {
		if u == id {
			c.ChannelUsers = append(c.ChannelUsers[:i], c.ChannelUsers[i+1:]...)
		}
	}

	Channels[c.ChannelName] = &Channel{
		ChannelName:  c.ChannelName,
		ChannelUsers: c.ChannelUsers,
	}

	// remove channel name from client channels
	for i, item := range Clients {
		if item.ID == id {
			for j, ch := range Clients[i].Channels {
				if ch == c.ChannelName {
					Clients[i].Channels = append(Clients[i].Channels[:j], Clients[i].Channels[j+1:]...)
				}
			}
		}
	}
}

func (c *Channel) FirstOrCreate(channelName string) error {

	if !channelExists(channelName) {

		newChannel := Channel{
			ChannelName:  channelName,
			ChannelUsers: []string{},
		}

		Channels[channelName] = &newChannel

	}

	c.ChannelName = Channels[channelName].ChannelName
	c.ChannelUsers = Channels[channelName].ChannelUsers
	
	if c.ChannelName == "" {
		return errors.New("channel not found")
	}

	return nil
}

func (c *Channel) InfoMessage() []byte {

	var joinedUsernames []string

	for _, id := range c.ChannelUsers {
		user := FindByID(id)

		if user.Username == "" {
			// 	user.Username = id
			panic("username not found")
		}

		joinedUsernames = append(joinedUsernames, user.Username)
	}

	info, err := json.Marshal(Message{
		Username:    "SERVER",
		ChannelName: c.ChannelName,
		Action:      "channel_info",
		Data: Channel{
			ChannelName:  c.ChannelName,
			ChannelUsers: joinedUsernames,
		},
	})

	if err != nil {
		log.Panicln(err)
	}
	return info
}

func (c *Channel) InChannel(id string) bool {
	for _, u := range c.ChannelUsers {
		if u == id {
			return true
		}
	}
	return false
}

func (c *Channel) Broadcast(message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return c.InChannel(q.Keys["id"].(string))
	})
}

func (c *Channel) BroadcastOther(s *melody.Session, message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return c.InChannel(q.Keys["id"].(string)) && q != s
	})
}

func ChannelFirst(channelName string) (c Channel, err error) {

	if channelExists(channelName) {

		return *Channels[channelName], nil
	}

	return c, errors.New("channel not found")
}

func channelExists(name string) bool {

	return Channels[name] != nil

}
