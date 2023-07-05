package chat

import (
	"chat2api/apireq"
	"chat2api/config"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
)

var (
	ErrNoAuth = `{
		"error": {
			"message": "You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth (i.e. Authorization: Bearer YOUR_KEY), or as the password field (with blank username) if you're accessing the API from your browser and are prompted for a username and password. You can obtain an API key from https://platform.openai.com/account/api-keys.",
			"type": "invalid_request_error",
			"param": null,
			"code": null
		}
	}`
	ErrKeyInvalid = `{
		"error": {
			"message": "Incorrect API key provided: sk-4yNZz***************************************6mjw. You can find your API key at https://platform.openai.com/account/api-keys.",
			"type": "invalid_request_error",
			"param": null,
			"code": "invalid_api_key"
		}
	}`
	ChatReqStr = `{
		"action": "next",
		"messages": [
			{
				"id": "aaa2a71f-eae4-4159-9efd-cd641985d50b",
				"author": {
					"role": "user"
				},
				"content": {
					"content_type": "text",
					"parts": [
						"hi"
					]
				},
				"metadata": {}
			}
		],
		"parent_message_id": "aaa10d6a-8671-4308-9886-8591990f5539",
		"model": "text-davinci-002-render-sha",
		"timezone_offset_min": -480,
		"history_and_training_disabled": false,
		"arkose_token": null
	}`
)

func Completions(r *ghttp.Request) {
	ctx := r.Context()
	g.Log().Debug(ctx, "Conversation start....................")
	authkey := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	if authkey == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrNoAuth))
		return
	}
	token := config.SK2TOKEN(ctx, authkey)
	if token == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrKeyInvalid))
		return
	}
	g.Log().Debug(ctx, "token: ", token)
	// 从请求中获取参数
	req := &apireq.Req{}
	err := r.GetRequestStruct(req)
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(gjson.New(`{"error": "bad request"}`))
		return
	}
	g.Dump(req)
	// 遍历 req.Messages 拼接 newMessages
	newMessages := ""
	for _, message := range req.Messages {
		newMessages += message.Content + "\n"
	}
	ChatReq := gjson.New(ChatReqStr)

	ChatReq.Set("messages.0.content.parts", newMessages)
	ChatReq.Set("messages.0.id", uuid.NewString())
	ChatReq.Set("parent_message_id", uuid.NewString())
	g.Dump(ChatReq)

}
