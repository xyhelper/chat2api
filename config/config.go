package config

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	PORT      = 8080
	APISERVER = "https://freechat.xyhelper.cn"
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
