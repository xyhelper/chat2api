package chat

import (
	"chat2api/apireq"
	"chat2api/apirespstream"
	"chat2api/config"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/launchdarkly/eventsource"
)

var (
	// client    = g.Client()
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
		"suggestions": [
			"Design a database schema for an online merch store.",
			"Make a content strategy for a newsletter featuring free local weekend events.",
			"Help me study vocabulary: write a sentence for me to fill in the blank, and I'll try to pick the correct option.",
			"Come up with 5 concepts for a retro-style arcade game."
		],
		"plugin_ids": [],
		"parent_message_id": "aaa10d6a-8671-4308-9886-8591990f5539",
		"model": "text-davinci-002-render-sha",
		"timezone_offset_min": -480,
		"history_and_training_disabled": true,
		"arkose_token": null,
		"force_paragen": false
	}`
	Chat4ReqStr = `
	{
		"action": "next",
		"messages": [
			{
				"id": "aaa2b182-d834-4f30-91f3-f791fa953204",
				"author": {
					"role": "user"
				},
				"content": {
					"content_type": "text",
					"parts": [
						"画一只猫1231231231"
					]
				},
				"metadata": {}
			}
		],
		"parent_message_id": "aaa11581-bceb-46c5-bc76-cb84be69725e",
		"model": "gpt-4-gizmo",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": true,
		"conversation_mode": {
			"gizmo": {
				"gizmo": {
					"id": "g-YyyyMT9XH",
					"organization_id": "org-OROoM5KiDq6bcfid37dQx4z4",
					"short_url": "g-YyyyMT9XH-chatgpt-classic",
					"author": {
						"user_id": "user-u7SVk5APwT622QC7DPe41GHJ",
						"display_name": "ChatGPT",
						"selected_display": "name",
						"is_verified": true
					},
					"voice": {
						"id": "ember"
					},
					"display": {
						"name": "ChatGPT Classic",
						"description": "The latest version of GPT-4 with no additional capabilities",
						"welcome_message": "Hello",
						"profile_picture_url": "https://files.oaiusercontent.com/file-i9IUxiJyRubSIOooY5XyfcmP?se=2123-10-13T01%3A11%3A31Z&sp=r&sv=2021-08-06&sr=b&rscc=max-age%3D31536000%2C%20immutable&rscd=attachment%3B%20filename%3Dgpt-4.jpg&sig=ZZP%2B7IWlgVpHrIdhD1C9wZqIvEPkTLfMIjx4PFezhfE%3D",
						"categories": []
					},
					"share_recipient": "link",
					"updated_at": "2023-11-06T01:11:32.191060+00:00",
					"last_interacted_at": "2023-11-18T07:50:19.340421+00:00",
					"tags": [
						"public",
						"first_party"
					]
				},
				"tools": [],
				"files": [],
				"product_features": {
					"attachments": {
						"type": "retrieval",
						"accepted_mime_types": [
							"text/x-c",
							"text/html",
							"application/x-latext",
							"text/plain",
							"text/x-ruby",
							"text/x-typescript",
							"text/x-c++",
							"text/x-java",
							"text/x-sh",
							"application/vnd.openxmlformats-officedocument.presentationml.presentation",
							"text/x-script.python",
							"text/javascript",
							"text/x-tex",
							"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
							"application/msword",
							"application/pdf",
							"text/x-php",
							"text/markdown",
							"application/json",
							"text/x-csharp"
						],
						"image_mime_types": [
							"image/jpeg",
							"image/png",
							"image/gif",
							"image/webp"
						],
						"can_accept_all_mime_types": true
					}
				}
			},
			"kind": "gizmo_interaction",
			"gizmo_id": "g-YyyyMT9XH"
		},
		"force_paragen": false,
		"force_rate_limit": false
	}`
	ApiRespStr = `{
		"id": "chatcmpl-LLKfuOEHqVW2AtHks7wAekyrnPAoj",
		"object": "chat.completion",
		"created": 1689864805,
		"model": "gpt-3.5-turbo",
		"usage": {
			"prompt_tokens": 0,
			"completion_tokens": 0,
			"total_tokens": 0
		},
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "Hello! How can I assist you today?"
				},
				"finish_reason": "stop",
				"index": 0
			}
		]
	}`
	ApiRespStrStream = `{
		"id": "chatcmpl-afUFyvbTa7259yNeDqaHRBQxH2PLH",
		"object": "chat.completion.chunk",
		"created": 1689867370,
		"model": "gpt-3.5-turbo",
		"choices": [
			{
				"delta": {
					"content": "Hello"
				},
				"index": 0,
				"finish_reason": null
			}
		]
	}`
	ApiRespStrStreamEnd = `{"id":"apirespid","object":"chat.completion.chunk","created":apicreated,"model": "apirespmodel","choices":[{"delta": {},"index": 0,"finish_reason": "stop"}]}`
)

func Completions(r *ghttp.Request) {
	ctx := r.Context()
	// g.Log().Debug(ctx, "Conversation start....................")
	if config.MAXTIME > 0 {
		// 创建带有超时的context
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(config.MAXTIME)*time.Second)
		defer cancel()
	}

	authkey := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	if authkey == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrNoAuth))
		return
	}
	// g.Log().Info(ctx, "authkey: ", authkey)
	var token string
	if config.PASSMODE {
		token = authkey
	} else {
		token = config.SK2TOKEN(ctx, authkey)
	}
	if token == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrKeyInvalid))
		return
	}
	// g.Log().Debug(ctx, "token: ", token)
	// 从请求中获取参数
	req := &apireq.Req{}
	err := r.GetRequestStruct(req)
	if err != nil {
		g.Log().Error(ctx, "r.GetRequestStruct(req) error: ", err)
		r.Response.Status = 400
		r.Response.WriteJson(gjson.New(`{"error": "bad request"}`))
		return
	}
	// g.Dump(req)
	// 遍历 req.Messages 拼接 newMessages
	newMessages := ""
	for _, message := range req.Messages {
		newMessages += message.Content + "\n"
	}
	// g.Dump(newMessages)
	var ChatReq *gjson.Json
	if gstr.HasPrefix(req.Model, "gpt-4") {
		ChatReq = gjson.New(Chat4ReqStr)
	} else {
		ChatReq = gjson.New(ChatReqStr)
	}

	ChatReq.Set("messages.0.content.parts.0", newMessages)
	ChatReq.Set("messages.0.id", uuid.NewString())
	ChatReq.Set("parent_message_id", uuid.NewString())
	if len(req.PluginIds) > 0 {
		ChatReq.Set("plugin_ids", req.PluginIds)
	}
	if config.KEEPHISTORY {
		ChatReq.Set("history_and_training_disabled", false)
	}
	// ChatReq.Dump()
	// 请求openai
	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
		"authkey":       config.AUTHKEY,
	}).Post(ctx, config.APISERVER, ChatReq.MustToJson())
	if err != nil {
		r.Response.Status = 500
		r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
		return
	}
	defer resp.Close()
	// defer resp.Body.Close()
	// 如果返回结果不是200
	if resp.StatusCode != 200 {
		g.Log().Error(ctx, "resp.StatusCode: ", resp.StatusCode)
		r.Response.Status = resp.StatusCode
		r.Response.WriteJson(gjson.New(resp.ReadAllString()))
		return
	}
	if resp.Header.Get("Content-Type") != "text/event-stream; charset=utf-8" && resp.Header.Get("Content-Type") != "text/event-stream" {
		g.Log().Error(ctx, "resp.Header.Get(Content-Type): ", resp.Header.Get("Content-Type"))
		r.Response.Status = 500
		r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
		return
	}

	// 流式返回
	if req.Stream {
		r.Response.Header().Set("Content-Type", "text/event-stream")
		r.Response.Header().Set("Cache-Control", "no-cache")
		r.Response.Header().Set("Connection", "keep-alive")
		// r.Response.Flush()
		message := ""
		decoder := eventsource.NewDecoder(resp.Body)
		defer decoder.Decode()

		id := config.GenerateID(29)
		for {
			event, err := decoder.Decode()
			if err != nil {
				// if err == io.EOF {
				// 	break
				// }
				// g.Log().Info(ctx, "释放资源")
				break
			}
			text := event.Data()
			// g.Log().Debug(ctx, "text: ", text)
			if text == "" {
				continue
			}
			if text == "[DONE]" {
				apiRespStrEnd := gstr.Replace(ApiRespStrStreamEnd, "apirespid", id)
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apicreated", gconv.String(time.Now().Unix()))
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apirespmodel", req.Model)
				r.Response.Writefln("data: " + apiRespStrEnd + "\n\n")
				r.Response.Flush()
				r.Response.Writefln("data: " + text + "\n\n")
				r.Response.Flush()
				continue
				// resp.Close()

				// break
			}
			// gjson.New(text).Dump()
			role := gjson.New(text).Get("message.author.role").String()
			if role == "assistant" {
				messageTemp := gjson.New(text).Get("message.content.parts.0").String()
				// g.Log().Debug(ctx, "messageTemp: ", messageTemp)
				// 如果 messageTemp 不包含 message 且plugin_ids为空
				if !gstr.Contains(messageTemp, message) && len(req.PluginIds) == 0 {
					continue
				}

				content := strings.Replace(messageTemp, message, "", 1)
				if content == "" {
					continue
				}
				message = messageTemp
				apiResp := gjson.New(ApiRespStrStream)
				apiResp.Set("id", id)
				apiResp.Set("created", time.Now().Unix())
				apiResp.Set("choices.0.delta.content", content)
				// if req.Model == "gpt-4" {
				// 	apiResp.Set("model", "gpt-4")
				// }
				apiResp.Set("model", req.Model)
				apiRespStruct := &apirespstream.ApiRespStreamStruct{}
				gconv.Struct(apiResp, apiRespStruct)
				// g.Dump(apiRespStruct)
				// 创建一个jsoniter的Encoder
				json := jsoniter.ConfigCompatibleWithStandardLibrary

				// 将结构体转换为JSON文本并保持顺序
				sortJson, err := json.Marshal(apiRespStruct)
				if err != nil {
					fmt.Println("转换JSON出错:", err)
					continue
				}
				r.Response.Writeln("data: " + string(sortJson) + "\n\n")
				r.Response.Flush()
			}

		}

	} else {
		// 非流式回应
		content := ""
		decoder := eventsource.NewDecoder(resp.Body)
		defer decoder.Decode()

		for {
			event, err := decoder.Decode()
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}
			text := event.Data()
			if text == "" {
				continue
			}
			if text == "[DONE]" {
				resp.Close()
				break
			}
			// gjson.New(text).Dump()
			role := gjson.New(text).Get("message.author.role").String()
			if role == "assistant" {
				message := gjson.New(text).Get("message.content.parts.0").String()
				if message != "" {
					content = message
				}
			}
		}
		completionTokens := CountTokens(content)
		promptTokens := CountTokens(newMessages)
		totalTokens := completionTokens + promptTokens

		apiResp := gjson.New(ApiRespStr)
		apiResp.Set("choices.0.message.content", content)
		id := config.GenerateID(29)
		apiResp.Set("id", id)
		apiResp.Set("created", time.Now().Unix())
		// if req.Model == "gpt-4" {
		// 	apiResp.Set("model", "gpt-4")
		// }
		apiResp.Set("model", req.Model)

		apiResp.Set("usage.prompt_tokens", promptTokens)
		apiResp.Set("usage.completion_tokens", completionTokens)
		apiResp.Set("usage.total_tokens", totalTokens)
		r.Response.WriteJson(apiResp)
	}

}
