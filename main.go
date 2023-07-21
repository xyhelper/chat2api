package main

import (
	"chat2api/config"
	"chat2api/v1/chat"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	s := g.Server()
	s.SetPort(config.PORT)
	chatGroup := s.Group("/v1/chat")
	chatGroup.ALL("/completions", chat.Completions)
	s.Run()
}
