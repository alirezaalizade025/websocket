package models

import (
	"errors"
	// "log"
	"time"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

type Client struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	ConnectAt string   `json:"connect_at"`
	Avatar    string   `json:"avatar"`
	Channels  []string `json:"channels"`
}

var Clients = map[string]*Client{}

// GenerateID generates a unique ID for the client.
// It assigns the generated ID to the client's ID field.
// If the generated ID is already used by another client, it recursively calls itself to generate a new ID.
// Returns an error if there is an issue generating the ID.
func GenerateID() string {

	id := uuid.NewString()

	for _, client := range Clients {
		if client.ID == id {
			return GenerateID()
		}
	}

	return id
}

// NewClient creates a new instance of the Client struct.
// It generates a unique ID for the client and sets the ConnectAt field to the current timestamp.
// The client is then added to the Clients slice.
// Returns the newly created client.
func NewClient() Client {
	client := Client{
		ID:        GenerateID(),
		ConnectAt: time.Now().Format(time.RFC3339), // Set ConnectAt to current timestamp
	}

	// Clients = append(Clients, client)

	Clients[client.ID] = &client

	return client
}

func (client Client) LeaveAllChannels(m *melody.Melody) {
	clientChannels := client.Channels
	for _, channelName := range clientChannels {
		channel, err := ChannelFirst(channelName)
		if err != nil {
			continue
		}

		channel.Leave(client.ID)

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
}

func (client Client) ActiveAllChannels(m *melody.Melody, s *melody.Session) {
	clientChannels := client.Channels
	for _, channelName := range clientChannels {
		channel, err := ChannelFirst(channelName)
		if err != nil {
			continue
		}

		channel.ActiveClient(client.ID)

		// ---- channel info broadcast -----
		channel.BroadcastOther(s, channel.InfoMessage(), m)
	}
}

func (client Client) InactiveAllChannels(m *melody.Melody, s *melody.Session) {
	clientChannels := client.Channels
	for _, channelName := range clientChannels {
		channel, err := ChannelFirst(channelName)
		if err != nil {
			continue
		}

		channel.InactiveClient(client.ID)

		// ---- channel info broadcast -----
		channel.BroadcastOther(s, channel.InfoMessage(), m)
	}
}


func MatchUsernameWithID(id, username string) error {

	if username == "" {
		return nil
	}

	if _, ok := Clients[id]; ok {
		Clients[id].Username = username
		return nil
	}

	return errors.New("Client not found")
}

func FindByUsername(username string) Client {
	for _, client := range Clients {
		if client.Username == username {
			return *client
		}
	}

	return Client{}
}

func FindByID(id string) Client {

	if client, ok := Clients[id]; ok {
		return *client
	}

	return Client{}
}
