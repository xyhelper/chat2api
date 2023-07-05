package main

import (
	"chat2api/config"
	"chat2api/v1/chat"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	s := g.Server()
	s.SetPort(config.PORT)
	g.Log().Info(ctx, config.SK2TOKEN(ctx, "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"))
	chatGroup := s.Group("/v1/chat")
	chatGroup.ALL("/completions", chat.Completions)
	s.Run()
}
