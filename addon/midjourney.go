package addon

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/pplorins/go-mitmproxy/dao/redis"
	"github.com/pplorins/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
	"gitlab.com/pplorins/wechat-official-accounts/chatgpt/shared"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// decode content-encoding then respond to client

var PROMPT_PATTERN = `^(?P<p_origin>.*) --seed (?P<seed>\d+)$`
var rePrompt = regexp.MustCompile(PROMPT_PATTERN)

type MidJourney struct {
	proxy.BaseAddon
}

func NewMidJourney() *MidJourney {
	o := &MidJourney{}
	o.Initialize()
	return o
}

func (o *MidJourney) Initialize() {
	redis.InitializeRedis()
}

//func (o *MidJourney) Response(f *proxy.Flow) {
//	if !o.isInteractionRPC(f) {
//		return
//	}
//
//	fmt.Println("xxx")
//}

func (o *MidJourney) Request(f *proxy.Flow) {
	if !o.isInteractionRPC(f) {
		return
	}

	bytes := f.Request.Body
	req := &shared.ImagineRequest{}
	e := json.Unmarshal(bytes, req)
	if e != nil {
		log.Error("[MidJourney plugin] unmarshal conversation request failed:%+v", e)
		return
	}

	if !o.isAnchorRequest(req) {
		return
	}

	seed := ""
	for _, op := range req.Data.Options {
		if op.Name != "prompt" {
			continue
		}
		prompt := op.Value
		seed, e = o.extractSeed(prompt)
		if e != nil {
			log.Error("[midJourney plugin] extract seed failed:%s", e)
			return
		}
		break
	}

	//push to redis
	go func() {
		hBytes, e := json.Marshal(f.Request.Header)
		if e != nil {
			log.Error("[MidJourney plugin] marshal header failed:%+v", e)
		}

		log.Infof("[MidJourney plugin] save anchor mj prompt http ctx to redis now")

		nowStr := time.Now().Format(shared.CLIENT_TIME_FORMAT)
		bc := &shared.MidJourneyBaseHttpRequestContext{
			Header: string(hBytes),
			UTime:  nowStr,
		}
		ir := &shared.ImagineRequestRedis{
			Req:   req,
			UTime: nowStr,
		}
		ls := &shared.MidJourneyLastSeedRedis{
			Seed:  seed,
			UTime: nowStr,
		}

		rc := redis.NewRedisClient()
		err := rc.WriteMidJourneyRequestHttpContext(context.Background(), seed, bc, ir, ls)
		if err != nil {
			log.Error("[MidJourney plugin] write MJ req-http-ctx to redis failed:%+v", e)
			return
		}
	}()
}

func (o *MidJourney) isInteractionRPC(f *proxy.Flow) bool {
	u := f.Request.URL
	if !strings.Contains(u.Hostname(), "discord.com") {
		return false
	}

	if f.Request.Method != http.MethodPost {
		return false
	}

	if !strings.HasSuffix(u.Path, "api/v9/interactions") {
		return false
	}

	return true
}

func (o *MidJourney) isAnchorRequest(req *shared.ImagineRequest) bool {
	for _, op := range req.Data.Options {
		if op.Name != "prompt" {
			continue
		}
		prompt := op.Value
		if prompt == shared.ANCHOR_PROMPT {
			return true
		}
	}
	return false
}

func (o *MidJourney) extractSeed(prompt string) (string, error) {
	ma := rePrompt.FindStringSubmatch(prompt)
	if len(ma) != 3 {
		return "", errors.Errorf("RE match prompt failed")
	}

	cgm := shared.GetCaptureGroupMap(rePrompt, ma)

	return cgm["seed"], nil
}
