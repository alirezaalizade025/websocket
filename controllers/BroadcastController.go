package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"socket/models"
	"socket/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/olahol/melody"
)

type BroadCastStoreRequest struct {
	Message   string   `json:"message" form:"message" binding:"omitempty,max=255"`
	Action    string   `json:"action" form:"action" binding:"required,max=50"`
	Data      string   `json:"data" form:"data" binding:"omitempty,max=1000"`
	Type      string   `json:"type" form:"type" binding:"omitempty,max=50"`
	Style     string   `json:"style" form:"style" binding:"omitempty,max=50"`
	AutoClose int      `json:"auto_close" form:"auto_close" binding:"omitempty,max=10000"`
	Receivers []string `json:"receivers" form:"receivers" binding:"omitempty"`
	Ignores   []string `json:"ignores" form:"ignores" binding:"omitempty"`
	SendMode  string   `json:"send_mode" form:"send_mode" binding:"omitempty,max=50"`
}

func Broadcast(c *gin.Context, m *melody.Melody) {

	// validation with gin
	request := &BroadCastStoreRequest{}
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(422, gin.H{
			"errors": translateError(err),
		})
		return
	}

	var validActions = []string{"message", "toast"}
	if !utils.Contains(validActions, request.Action) {
		c.JSON(422, gin.H{
			"errors": "Invalid action",
		})
		return
	}

	if request.SendMode == "" {
		request.SendMode = "all"
	} else {
		var validSendMode = []string{"active_only", "active_first", "all"}
		if !utils.Contains(validSendMode, request.SendMode) {
			c.JSON(422, gin.H{
				"errors": "Invalid send mode",
			})
			return
		}
	}

	var message []byte
	var err error

	if request.Action == "toast" {
		// generate message
		message, err = json.Marshal(models.Message{
			ChannelName: "API",
			Action:      request.Action,
			Data: map[string]interface{}{
				"message":    request.Message,
				"type":       request.Style,
				"auto_close": request.AutoClose,
			},
		})
		if err != nil {
			log.Println(err)
		}
	}

	// find receivers
	if len(request.Receivers) > 0 {

		receivers := request.Receivers

		// remove ignores
		if len(request.Ignores) > 0 {
			for _, ignore := range request.Ignores {
				clients := models.FindByUsername(ignore)

				for _, c := range clients {
					receivers = utils.Remove(receivers, c.Username)
				}
			}
		}

		var receiversIds []string

		for _, receiver := range receivers {
			clients := models.FindByUsername(receiver)

			receiversIds = handleActiveForReceivers(clients, request, receiversIds)
		}

		m.BroadcastFilter(message, func(s *melody.Session) bool {
			return utils.Contains(receiversIds, s.Keys["id"].(string))
		})

	} else {

		// remove ignores
		if len(request.Ignores) > 0 {

			var receiversIds []string

			sessions, err := m.Sessions()
			if err != nil {
				log.Println("broadcast ignore receivers:", err.Error())
				return
			}

			for _, session := range sessions {
				client := models.FindByID(session.Keys["id"].(string))

				// remove ignores
				if utils.Contains(request.Ignores, client.Username) {
					continue
				}

				clients := models.FindByUsername(client.Username)

				receiversIds = handleActiveForReceivers(clients, request, receiversIds)
			}

			m.BroadcastFilter(message, func(s *melody.Session) bool {
				return utils.Contains(receiversIds, s.Keys["id"].(string))
			})

		} else {

			var receiversIds []string

			sessions, err := m.Sessions()
			if err != nil {
				log.Println("broadcast ignore receivers:", err.Error())
				return
			}

			for _, session := range sessions {
				client := models.FindByID(session.Keys["id"].(string))

				clients := models.FindByUsername(client.Username)

				receiversIds = handleActiveForReceivers(clients, request, receiversIds)
			}

			m.BroadcastFilter(message, func(s *melody.Session) bool {
				return utils.Contains(receiversIds, s.Keys["id"].(string))
			})
		}

	}

}

func handleActiveForReceivers(clients []*models.Client, request *BroadCastStoreRequest, receiversIds []string) []string {
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

		clientReceiversIds := handleSendMode(request.SendMode, activeClientIds, inactiveClientIds)

		receiversIds = append(receiversIds, clientReceiversIds...)
	}
	return receiversIds
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

// translateError is a helper function to translate error from validator
func translateError(err error) (errTxt string) {
	validationErrors := err.(validator.ValidationErrors)
	for _, e := range validationErrors {
		errTxt = i18n(composeMsgID(e), e.Param())
		break
	}
	return
}

var tagPrefixMap = map[string]string{
	"required": "Required",
	"email":    "InvalidEmail",
	"min":      "ShouldMin",
	"max":      "ShouldMax",
	"len":      "ShouldLen",
	"eq":       "ShouldEq",
	"gt":       "ShouldGt",
	"gte":      "ShouldGte",
	"lt":       "ShouldLt",
	"lte":      "ShouldLte",
}

// i18n is a translation function
func i18n(msgID string, params ...interface{}) string {
	// implement the translation with msgID
	return msgID
}

// composeMsgID is a helper function to compose error message ID
func composeMsgID(e validator.FieldError) string {
	if prefix, ok := tagPrefixMap[e.Tag()]; ok {
		return fmt.Sprintf("%s %s", prefix, e.Field())
	}
	return ""
}
