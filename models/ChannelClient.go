package models

import "time"

type ChannelClient struct {
	ClientID  string `json:"client_id"`
	ChannelID int `json:"channel_id"`
	Status    Status `json:"status"`
	JoinAt    string `json:"join_at"`
}

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

var ChannelClients = []*ChannelClient{}

func Insert(ClientID string, ChannelID int) {

	// insert if not exists
	for _, v := range ChannelClients {
		if v.ClientID == ClientID && v.ChannelID == ChannelID {
			return
		}
	}

	ChannelClients = append(ChannelClients, &ChannelClient{
		ClientID:  ClientID,
		ChannelID: ChannelID,
		Status:    Active,
		JoinAt:    time.Now().Format(time.RFC3339),
	})

}

func DeleteByClientId(ClientId string) {
	for k, v := range ChannelClients {
		if v.ClientID == ClientId {
			ChannelClients = append(ChannelClients[:k], ChannelClients[k+1:]...)
		}
	}
}

func DeleteByChannelId(ChannelId int) {
	for k, v := range ChannelClients {
		if v.ChannelID == ChannelId {
			ChannelClients = append(ChannelClients[:k], ChannelClients[k+1:]...)
		}
	}
}

func UpdateStatus(ClientID string, ChannelID int, status Status) {
	for _, v := range ChannelClients {
		if v.ClientID == ClientID && v.ChannelID == ChannelID {
			v.Status = status
		}
	}
}

func GetChannelClients(ChannelID int) []ChannelClient {
	var clients []ChannelClient
	for _, v := range ChannelClients {
		if v.ChannelID == ChannelID {
			clients = append(clients, *v)
		}
	}
	return clients
}

func GetChannelClientsByStatus(ChannelID int, status Status) []ChannelClient {
	var clients []ChannelClient
	for _, v := range ChannelClients {
		if v.ChannelID == ChannelID && v.Status == status {
			clients = append(clients, *v)
		}
	}
	return clients
}
