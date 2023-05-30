package addon

import (
	"context"
	"encoding/json"
	"github.com/lqqyt2423/go-mitmproxy/dao/redis"
	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// decode content-encoding then respond to client

const (
	MJ_SESSION_ID = "f841f6b3-e854-4107-98d3-5ba81d9bfc65"
)

type ConversationRequest struct {
	Action   string `json:"action"`
	Messages []struct {
		ID     string `json:"id"`
		Author struct {
			Role string `json:"role"`
		} `json:"author"`
		Content struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
	} `json:"messages"`
	ConversationID             string `json:"conversation_id"`
	ParentMessageID            string `json:"parent_message_id"`
	Model                      string `json:"model"`
	TimezoneOffsetMin          int    `json:"timezone_offset_min"`
	HistoryAndTrainingDisabled bool   `json:"history_and_training_disabled"`
}

type ConversationResponse struct {
	Message struct {
		ID     string `json:"id"`
		Author struct {
			Role     string      `json:"role"`
			Name     interface{} `json:"name"`
			Metadata struct {
			} `json:"metadata"`
		} `json:"author"`
		CreateTime float64     `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		Status   string  `json:"status"`
		EndTurn  bool    `json:"end_turn"`
		Weight   float64 `json:"weight"`
		Metadata struct {
			MessageType   string `json:"message_type"`
			ModelSlug     string `json:"model_slug"`
			FinishDetails struct {
				Type string `json:"type"`
				Stop string `json:"stop"`
			} `json:"finish_details"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationID string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}

type OpenAI struct {
	proxy.BaseAddon
}

func NewOpenAI() *OpenAI {
	o := &OpenAI{}
	o.Initialize()
	return o
}

func (o *OpenAI) Initialize() {
	redis.InitializeRedis()
}

func (o *OpenAI) Response(f *proxy.Flow) {
	var e error
	defer func() {
		if e != nil {
			//todo: notify with
		}
	}()

	if !o.isConversationRPC(f) {
		return
	}

	req := &ConversationRequest{}
	e = json.Unmarshal(f.Request.Body, req)
	if e != nil {
		log.Error("[openAI plugin] unmarshal conversation request failed:%+v", e)
		return
	}

	c1 := req.ConversationID
	if c1 == MJ_SESSION_ID {
		return
	}

	var decodedBody []byte
	decodedBody, e = f.Response.DecodedBody()
	if e != nil {
		log.Error("[openAI plugin] get decoded body failed:%+v", e)
		return
	}

	s := string(decodedBody)
	events := strings.Split(s, "\n")
	if len(events) < 2 {
		e = errors.Errorf("[openAI plugin] cannot split decoded body")
		log.Error("[openAI plugin] get decoded body failed:%+v", e)
		return
	}

	//remove empty events
	filteredEvents := make([]string, 0, len(events))
	for _, ev := range events {
		if ev == "" {
			continue
		}
		filteredEvents = append(filteredEvents, ev)
	}

	lastEv := filteredEvents[len(filteredEvents)-2]
	lastEv = strings.TrimPrefix(lastEv, "data:")

	rsp := &ConversationResponse{}
	e = json.Unmarshal([]byte(lastEv), rsp)
	if e != nil {
		log.Error("[openAI plugin] unmarshal conversation response failed:%+v", e)
		return
	}

	c2 := rsp.ConversationID
	if c1 != c2 {
		e = errors.Errorf("[openAI plugin] req/rsp conversationID not match:%s|%s", c1, c2)
		log.Error("[openAI plugin] rsp msg check failed:%+v", e)
		return
	}

	msgs := rsp.Message.Content.Parts
	l := len(msgs)
	if l != 1 {
		e = errors.Errorf("[openAI plugin] rsp content parts length incorrect:%d", l)
		log.Error("[openAI plugin] rsp msg check failed:%+v", e)
		return
	}
	msg := msgs[0]

	//push to redis
	go func() {
		rc := redis.NewRedisClient()
		err := rc.NotifyAnswer(context.Background(), c2, msg)
		if err != nil {
			log.Error("[openAI plugin] notify gpt answer failed:%+v", e)
			return
		}
	}()
}

func (o *OpenAI) isConversationRPC(f *proxy.Flow) bool {
	u := f.Request.URL
	if !strings.Contains(u.Hostname(), "chat.openai.com") {
		return false
	}

	if f.Request.Method != http.MethodPost {
		return false
	}

	if !strings.HasSuffix(u.Path, "backend-api/conversation") {
		return false
	}

	return true
}
