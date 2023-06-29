package main

import (
	"fmt"
	"github.com/pplorins/go-mitmproxy/addon"
	cf "github.com/pplorins/go-mitmproxy/config"
	"github.com/pplorins/go-mitmproxy/proxy"
	"github.com/pplorins/go-mitmproxy/web"
	log "github.com/sirupsen/logrus"
	"io"
	rawLog "log"
	"os"
	"path"
	"strings"
)

//var once sync.Once
//var logger *logrus.Logger

const (
	LOG_TIME_FORMAT = "2006-01-02 15:04:05.999-07:00"
)

func main() {

	//a := os.Environ()
	//fmt.Println(a)

	config := cf.LoadConfig()

	if config.Debug > 0 {
		rawLog.SetFlags(rawLog.LstdFlags | rawLog.Lshortfile)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	//if config.Debug == 2 {
	log.SetReportCaller(true)
	//}
	f, e := os.OpenFile(path.Join("log", "go_mitmproxy.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	mw := io.MultiWriter(f, os.Stdout)
	log.SetOutput(mw)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  LOG_TIME_FORMAT,
		ForceColors:      true,
		DisableColors:    false,
		DisableTimestamp: false,
	})

	opts := &proxy.Options{
		Debug:             config.Debug,
		Addr:              config.Addr,
		StreamLargeBodies: 1024 * 1024 * 5,
		SslInsecure:       config.SslInsecure,
		CaRootPath:        config.CertPath,
		Upstream:          config.Upstream,
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	if config.Version {
		fmt.Println("go-mitmproxy: " + p.Version)
		os.Exit(0)
	}

	log.Infof("go-mitmproxy version %v\n", p.Version)

	if len(config.IgnoreHosts) > 0 {
		p.SetShouldInterceptRule(func(address string) bool {
			return !matchHost(address, config.IgnoreHosts)
		})
	}
	if len(config.AllowHosts) > 0 {
		p.SetShouldInterceptRule(func(address string) bool {
			return matchHost(address, config.AllowHosts)
		})
	}

	p.AddAddon(&proxy.LogAddon{})
	p.AddAddon(web.NewWebAddon(config.WebAddr))

	if config.MapRemote != "" {
		mapRemote, err := addon.NewMapRemoteFromFile(config.MapRemote)
		if err != nil {
			log.Warnf("load map remote error: %v", err)
		} else {
			p.AddAddon(mapRemote)
		}
	}

	if config.MapLocal != "" {
		mapLocal, err := addon.NewMapLocalFromFile(config.MapLocal)
		if err != nil {
			log.Warnf("load map local error: %v", err)
		} else {
			p.AddAddon(mapLocal)
		}
	}

	if config.Dump != "" {
		dumper := addon.NewDumperWithFilename(config.Dump, config.DumpLevel)
		p.AddAddon(dumper)
	}

	p.AddAddon(addon.NewOpenAI())
	p.AddAddon(addon.NewMidJourney())

	log.Fatal(p.Start())
}

func matchHost(address string, hosts []string) bool {
	hostname, port := splitHostPort(address)
	for _, host := range hosts {
		h, p := splitHostPort(host)
		if matchHostname(hostname, h) && (p == "" || p == port) {
			return true
		}
	}
	return false
}

func matchHostname(hostname string, h string) bool {
	if h == "*" {
		return true
	}
	if strings.HasPrefix(h, "*.") {
		return hostname == h[2:] || strings.HasSuffix(hostname, h[1:])
	}
	return h == hostname
}

func splitHostPort(address string) (string, string) {
	index := strings.LastIndex(address, ":")
	if index == -1 {
		return address, ""
	}
	return address[:index], address[index+1:]
}

//func InitLog() {
//	once.Do(func() {
//		logger = logrus.New()
//		logger.SetReportCaller(true)
//		logger.SetFormatter(&prefixed.TextFormatter{
//			FullTimestamp:    true,
//			TimestampFormat:  LOG_TIME_FORMAT,
//			ForceFormatting:  true,
//			ForceColors:      true,
//			DisableColors:    false,
//			DisableTimestamp: false,
//		})
//		logger.SetLevel(logrus.InfoLevel)
//		f, e := os.OpenFile(path.Join("log", "biz.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//		if e != nil {
//			panic(e)
//		}
//		mw := io.MultiWriter(f, os.Stdout)
//		logger.SetOutput(mw)
//	})
//
//	logger.Infof("logger initialized successfully")
//}
