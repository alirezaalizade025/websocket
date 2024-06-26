package models

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

type Client struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	ConnectAt time.Time  `json:"connect_at"`
	Avatar    string     `json:"avatar"`
	PingAt    *time.Time `json:"ping_at"`
	Status    Status     `json:"status"`
}

var Clients = &sync.Map{}

var AdminClient = Client{}

// GenerateID generates a unique ID for the client.
// It assigns the generated ID to the client's ID field.
// If the generated ID is already used by another client, it recursively calls itself to generate a new ID.
// Returns an error if there is an issue generating the ID.
func GenerateID() string {

	id := uuid.NewString()

	if _, ok := Clients.Load(id); ok {
		return GenerateID()
	}

	return id
}

// NewClient creates a new instance of the Client struct.
// It generates a unique ID for the client and sets the ConnectAt field to the current timestamp.
// The client is then added to the Clients slice.
// Returns the newly created client.
func NewClient() *Client {
	client := &Client{
		ID:        GenerateID(),
		ConnectAt: time.Now(), // Set ConnectAt to current timestamp
		Status:    Active,
	}

	Clients.Store(client.ID, client)

	return client
}

func (client *Client) GetSession(m *melody.Melody) (*melody.Session, error) {
	sessions, err := m.Sessions()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, session := range sessions {
		if session.Keys["id"] == client.ID {

			return session, nil
		}
	}

	return nil, errors.New("session not found")
}

func (client *Client) Delete() {
	Clients.Delete(client.ID)
}

func (client *Client) UpdateClient(data Message) {

	if data.Data == "" {
		return
	}

	clientData := data.Data.(map[string]interface{})

	if clientData["username"] != nil {
		client.Username = clientData["username"].(string)
	}

	if clientData["avatar"] != nil {
		client.Avatar = clientData["avatar"].(string)
	}

	Clients.Store(client.ID, client)
}

func (client *Client) InitAdmin() {

	AdminClient = Client{
		ID:       client.ID,
		Username: "Admin",
	}
}

func (c *Client) InitMessage() []byte {

	info, err := json.Marshal(map[string]string{
		"action": "client_initd",
	})

	if err != nil {
		log.Panicln(err)
	}

	return info
}

func (c *Client) AdminInitMessage() []byte {

	info, err := json.Marshal(map[string]string{
		"action": "admin_initd",
	})

	if err != nil {
		log.Panicln(err)
	}

	return info
}

func (client Client) LeaveAllChannels(m *melody.Melody) {

	for _, item := range ChannelClients {
		if item.ClientID == client.ID {

			channel, err := ChannelFirstById(item.ChannelID)

			if err != nil {
				log.Println("LeaveAllChannels: " + err.Error())
				continue
			}

			channel.Leave(client.ID)

			// ---- channel info broadcast -----
			channel.BroadcastOther(nil, channel.InfoMessage(), m)
		}
	}
}

func ClientsInfo() []byte {

	joinedUsernames := []map[string]interface{}{}

	Clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)

		if client.Username == "" {
			// 	user.Username = id
			Clients.Delete(client.ID)

			return true
		}

		// if username not exists in joinedUsernames, append it
		for _, item := range joinedUsernames {
			if item["username"] == client.Username {

				// if existed item status is inactive and current client status is active, update status
				if item["status"] == Inactive && client.Status == Active {
					item["status"] = Active
				}

				return true
			}
		}

		joinedUsernames = append(joinedUsernames, map[string]interface{}{
			"username": client.Username,
			"avatar":   client.Avatar,
			"status":   client.Status,
		})

		return true
	})

	type Data struct {
		// ChannelName    string                   `json:"channel_name"`
		ChannelClients []map[string]interface{} `json:"users"`
	}

	ChannelName := MessageTypeChannelGeneral

	info, err := json.Marshal(Message{
		ChannelName: (*string)(&ChannelName),
		Action:      "clients_info",
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

func (client Client) ActiveAllChannels(m *melody.Melody, s *melody.Session) {

	for i, item := range ChannelClients {
		if item.ClientID == client.ID {

			channel, err := ChannelFirstById(item.ChannelID)

			if err != nil {
				log.Println("ActiveAllChannels: " + err.Error())
			}

			ChannelClients[i].Status = Active

			// ---- channel info broadcast -----
			channel.BroadcastOther(s, channel.InfoMessage(), m)
		}
	}
}

func (client Client) InactiveAllChannels(m *melody.Melody, s *melody.Session) {
	for i, item := range ChannelClients {
		if item.ClientID == client.ID {

			ChannelClients[i].Status = Inactive

			channel, err := ChannelFirstById(item.ChannelID)

			if err != nil {
				log.Println("InactiveAllChannels: " + err.Error())
			}

			// ---- channel info broadcast -----
			channel.BroadcastOther(s, channel.InfoMessage(), m)
		}
	}
}

func FindByUsername(username string) (c []*Client) {

	Clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)

		if client.Username == username {
			c = append(c, client)
		}

		return true
	})

	return
}

func FindByID(id string) Client {

	if client, ok := Clients.Load(id); ok {
		return *client.(*Client)
	}

	return Client{}
}

func (client *Client) SetStatus(status Status) {
	client.Status = status

	Clients.Store(client.ID, client)
}
