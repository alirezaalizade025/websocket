package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"socket/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/olahol/melody"
)

type BroadCastStoreRequest struct {
	ChannelName string `json:"channel_name" form:"channel_name" binding:"omitempty,max=255"`
	Message     string `json:"message" form:"message" binding:"omitempty,max=255"`
	Action      string `json:"action" form:"action" binding:"omitempty,max=50"`
	Data        string `json:"data" form:"data" binding:"omitempty,max=1000"`
	Type        string `json:"type" form:"type" binding:"omitempty,max=50"`
	AutoClose   int    `json:"auto_close" form:"auto_close" binding:"omitempty,max=10000"`
}

func Broadcast(c *gin.Context, m *melody.Melody) {

	// validation with gin
	request := &BroadCastStoreRequest{
		ChannelName: c.PostForm("channel_name"),
		Message:     c.PostForm("message"),
		Action:      c.PostForm("action"), // todo: null check
		Data:        c.PostForm("data"),
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(422, gin.H{
			"errors": translateError(err),
		})

		return
	}

	// find channel
	if request.ChannelName != "" {
		channel := models.Channel{}
		err := channel.FirstOrCreate(request.ChannelName)
		if err != nil {
			c.JSON(422, gin.H{
				"errors": "Channel not found",
			})
			return
		}
	}

	// generate message
	message, err := json.Marshal(models.Message{
		ChannelName: request.ChannelName,
		Action:      request.Action,
		Data: map[string]interface{}{
			"message":    request.Message,
			"type":       request.Type,
			"auto_close": request.AutoClose,
		},
	})

	if err != nil {
		log.Panicln(err)
	}

	m.Broadcast(message)

	// channel.Broadcast(message, m)
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
