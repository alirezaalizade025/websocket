package models

import (
	"errors"
	// "log"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	ConnectAt string   `json:"connect_at"`
	Channels  []string `json:"channels"`
}

var Clients = []Client{}

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

	Clients = append(Clients, client)

	return client
}

func MatchUsernameWithID(id, username string) error {

	if username == "" {
		return nil
	}

	for i, client := range Clients {
		if client.ID == id {
			Clients[i].Username = username

			return nil
		}
	}

	return errors.New("Client not found")
}

func FindByUsername(username string) Client {
	for _, client := range Clients {
		if client.Username == username {
			return client
		}
	}

	return Client{}
}

func FindByID(id string) Client {
	// log.Println(Clients)
	for _, client := range Clients {

		if client.ID == id {

			return client
		} else {
			// log.Println(client.ID, id)
		}
	}

	return Client{}
}