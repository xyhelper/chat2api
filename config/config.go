package config

import (
	"math/rand"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
)

var (
	PORT        = 8080
	APISERVER   = "http://chatproxy/backend-api/conversation"
	APIHOST     = "http://chatproxy"
	PASSMODE    = false
	MAXTIME     = 0
	NOPLUGINS   = false
	KEEPHISTORY = false
	AUTHKEY     = ""
)

func init() {
	ctx := gctx.GetInitCtx()
	port := g.Cfg().MustGetWithEnv(ctx, "PORT").Int()
	if port != 0 {
		PORT = port
	}
	apiServer := g.Cfg().MustGetWithEnv(ctx, "APISERVER").String()
	if apiServer != "" {
		APISERVER = apiServer
	}
	// 从apiServer中获取APIHOST
	apihost := gstr.SubStr(apiServer, 0, gstr.PosR(apiServer, "/backend-api/conversation"))
	if apihost != "" {
		APIHOST = apihost
	}
	passMode := g.Cfg().MustGetWithEnv(ctx, "PASSMODE").Bool()
	if passMode {
		PASSMODE = passMode
	}
	maxtime := g.Cfg().MustGetWithEnv(ctx, "MAXTIME").Int()
	if maxtime > 0 {
		MAXTIME = maxtime
	}
	noplugins := g.Cfg().MustGetWithEnv(ctx, "NOPLUGINS").Bool()
	if noplugins {
		NOPLUGINS = noplugins
	}
	keepHistory := g.Cfg().MustGetWithEnv(ctx, "KEEPHISTORY").Bool()
	if keepHistory {
		KEEPHISTORY = keepHistory
	}
	authKey := g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String()
	if authKey != "" {
		AUTHKEY = authKey
	}

	g.Log().Info(ctx, "PORT:", PORT)
	g.Log().Info(ctx, "APISERVER:", APISERVER)
	g.Log().Info(ctx, "PASSMODE:", PASSMODE)
	g.Log().Info(ctx, "APIHOST:", APIHOST)
	g.Log().Info(ctx, "MAXTIME:", MAXTIME)
	g.Log().Info(ctx, "NOPLUGINS:", NOPLUGINS)
	g.Log().Info(ctx, "KEEPHISTORY:", KEEPHISTORY)
	g.Log().Info(ctx, "AUTHKEY:", AUTHKEY)
}

func SK2TOKEN(ctx g.Ctx, SK string) (token string) {
	// 检查SK是否有效格式 如 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
	// 如果有效则返回token
	// 如果无效则返回空字符串
	sampleKey := "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"
	if len(SK) != len(sampleKey) {
		return ""
	}

	return g.Cfg().MustGetWithEnv(ctx, SK).String()
}

func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// rand.Seed(time.Now().UnixNano())

	id := "chatcmpl-"
	for i := 0; i < length; i++ {
		id += string(charset[rand.Intn(len(charset))])
	}
	return id
}
