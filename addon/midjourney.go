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

func (o *MidJourney) parseImagineJson(f *proxy.Flow, simplifiedFormat bool) (string, error) {
	req := f.Request
	if simplifiedFormat {
		return string(req.Body), nil
	}

	//parse boundary, original format is not standard multipul-part/form-data
	s := string(req.Body)
	target := `name="payload_json"`
	start := strings.Index(s, target)
	if start == -1 {
		return "", errors.Errorf("no start boundary data found")
	}
	start += len(target)

	s = s[start:]

	end := strings.Index(s, "------")
	if end == -1 {
		return "", errors.Errorf("no end boundary data found")
	}
	//skip over "\r\n\r\n", which has 4 bytes
	s = s[4:end]

	return s, nil
}

func (o *MidJourney) check(f *proxy.Flow) (bool, bool) {
	hdr := f.Request.Header

	defer func() {
		delete(f.Request.Header, shared.HTTP_HDR_JOB_TYPE)
	}()

	//requests without this key
	jtl, ok := hdr[shared.HTTP_HDR_JOB_TYPE]
	if !ok {
		return true, false //, shared.MJ_JOB_WI
	}
	if len(jtl) < 1 {
		return true, false //, shared.MJ_JOB_WI
	}

	//jobType := jtl[0]
	//if jobType == shared.MJ_JOB_WI {
	//	return true, false//, jobType
	//}
	//
	//if jobType == shared.MJ_JOB_I {
	//	return true, true, jobType
	//}

	//program generated requests no need to process
	return false, false //, jobType
}

func (o *MidJourney) Request(f *proxy.Flow) {
	if !o.isInteractionRPC(f) {
		return
	}

	needProcess, simplifiedFormat := o.check(f)
	if !needProcess {
		return
	}

	j, e := o.parseImagineJson(f, simplifiedFormat)
	if e != nil {
		//uvr reqeust could also reach here
		log.Warnf("[MidJourney plugin] parse imagine payload failed:%+v, maybe a uvr request", e)
		return
	}

	req := &shared.InteractionRequest{}
	e = json.Unmarshal([]byte(j), req)
	if e != nil {
		log.Errorf("[MidJourney plugin] unmarshal conversation request failed:%+v", e)
		return
	}

	ctx := context.Background()
	now := time.Now()
	nowStr := now.Format(shared.CLIENT_TIME_FORMAT)

	//describe cmd
	if req.Data.Name == "describe" {
		go func() {
			ir := &shared.InteractionRequestRedis{
				Req:   req,
				UTime: nowStr,
			}
			r := redis.NewRedisClient()
			e := r.WriteMJDescReqDetail(ctx, ir)
			if e != nil {
				log.Error("[MidJourney plugin] write MJ describe req-http-ctx to redis failed:%+v", e)
				return
			}
		}()
		return
	}

	if req.Data.Name == "blend" {
		go func() {
			ir := &shared.InteractionRequestRedis{
				Req:   req,
				UTime: nowStr,
			}
			r := redis.NewRedisClient()
			e := r.WriteMJBlendReqDetail(ctx, ir)
			if e != nil {
				log.Error("[MidJourney plugin] write MJ describe req-http-ctx to redis failed:%+v", e)
				return
			}
		}()
		return
	}

	//imagine cmd
	taskID := ""
	for _, op := range req.Data.Options {
		if op.Name != "prompt" {
			continue
		}
		prompt := op.Value
		p, e := prompt.MarshalJSON()
		if e != nil {
			log.Error("[MidJourney plugin] marshal data.Options.Value failed:%+v", e)
			return
		}
		ps := strings.Trim(string(p), "\"")
		taskID, e = o.extractSeed(ps)
		if e != nil {
			log.Warnf("[midJourney plugin] extract taskID failed:%s,mayBe webImagine without seed,ignore & continue..", e)
			taskID = ""
		}
		break
	}

	go func() {
		hBytes, e := json.Marshal(f.Request.Header)
		if e != nil {
			log.Error("[MidJourney plugin] marshal header failed:%+v", e)
		}

		log.Infof("[MidJourney plugin] save anchor mj prompt http ctx to redis now")

		bc := &shared.MidJourneyBaseHttpRequestContext{
			Header: string(hBytes),
			UTime:  nowStr,
		}
		ir := &shared.InteractionRequestRedis{
			Req:   req,
			UTime: nowStr,
		}

		var ls *shared.MidJourneyLastTaskRedis = nil
		if taskID != "" {
			ls = &shared.MidJourneyLastTaskRedis{
				TaskID:  taskID,
				UTime:   nowStr,
				UTimeTS: uint32(now.Unix()),
			}
		}

		//save to redis
		r := redis.NewRedisClient()
		e = r.WriteMidJourneyRequestHttpContext(ctx, taskID, bc, ir, ls)
		if e != nil {
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

func (o *MidJourney) extractSeed(prompt string) (string, error) {
	ma := rePrompt.FindStringSubmatch(prompt)
	if len(ma) != 3 {
		return "", errors.Errorf("RE match prompt failed")
	}

	cgm := shared.GetCaptureGroupMap(rePrompt, ma)

	return cgm["seed"], nil
}
