package chat

import (
	"chat2api/apireq"
	apirespstream "chat2api/apirespStream"
	"chat2api/config"
	"fmt"
	"io"
	"net/http"
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
		"history_and_training_disabled": true,
		"arkose_token": null
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

	authkey := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	if authkey == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrNoAuth))
		return
	}
	g.Log().Info(ctx, "authkey: ", authkey)
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
	// g.Dump(req)
	// 遍历 req.Messages 拼接 newMessages
	newMessages := ""
	for _, message := range req.Messages {
		newMessages += message.Content + "\n"
	}
	ChatReq := gjson.New(ChatReqStr)

	ChatReq.Set("messages.0.content.parts.0", newMessages)
	ChatReq.Set("messages.0.id", uuid.NewString())
	ChatReq.Set("parent_message_id", uuid.NewString())
	if req.Model == "gpt-4" {
		ChatReq.Set("model", "gpt-4-plugins")
	}

	// 请求openai
	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}).Post(ctx, config.APISERVER, gjson.New(ChatReq).MustToJson())
	if err != nil {
		r.Response.Status = 500
		r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
		return
	}
	defer resp.Close()
	// resp.RawDump()
	// 如果返回结果不是200
	if resp.StatusCode != 200 {
		r.Response.Status = resp.StatusCode
		r.Response.WriteJson(gjson.New(resp.ReadAllString()))
		return
	}

	// 流式返回
	if req.Stream {
		//  流式回应
		rw := r.Response.RawWriter()
		flusher, ok := rw.(http.Flusher)
		if !ok {
			g.Log().Error(ctx, "rw.(http.Flusher) error")
			r.Response.WriteStatusExit(500)
			return
		}
		r.Response.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		r.Response.Header().Set("Cache-Control", "no-cache")
		r.Response.Header().Set("Connection", "keep-alive")
		// r.Response.Flush()
		message := ""
		decoder := eventsource.NewDecoder(resp.Body)
		id := config.GenerateID(29)
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
				apiRespStrEnd := gstr.Replace(ApiRespStrStreamEnd, "apirespid", id)
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apicreated", gconv.String(time.Now().Unix()))
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apirespmodel", req.Model)
				rw.Write([]byte("data: " + apiRespStrEnd + "\n\n"))
				// apiRespStrEnd.Set("id", id)
				// apiRespStrEnd.Set("created", time.Now().Unix())
				// if req.Model == "gpt-4" {
				// 	apiRespStrEnd.Set("model", "gpt-4")
				// }
				// rw.Write([]byte("data: " + apiRespStrEnd.String() + "\n\n"))
				rw.Write([]byte("data: " + text + "\n\n"))
				flusher.Flush()
				break
			}
			// gjson.New(text).Dump()
			role := gjson.New(text).Get("message.author.role").String()
			if role == "assistant" {
				messageTemp := gjson.New(text).Get("message.content.parts.0").String()
				//
				content := strings.Replace(messageTemp, message, "", 1)
				if content == "" {
					continue
				}
				message = messageTemp
				apiResp := gjson.New(ApiRespStrStream)
				apiResp.Set("id", id)
				apiResp.Set("created", time.Now().Unix())
				apiResp.Set("choices.0.delta.content", content)
				if req.Model == "gpt-4" {
					apiResp.Set("model", "gpt-4")
				}
				apiRespStruct := &apirespstream.ApiRespStream{}
				gconv.Struct(apiResp, apiRespStruct)
				// g.Dump(apiRespStruct)
				// 创建一个jsoniter的Encoder
				json := jsoniter.ConfigCompatibleWithStandardLibrary

				// 将结构体转换为JSON文本并保持顺序
				sortJson, err := json.Marshal(apiRespStruct)
				if err != nil {
					fmt.Println("转换JSON出错:", err)
					return
				}
				rw.Write([]byte("data: " + string(sortJson) + "\n\n"))
				flusher.Flush()
			}

		}

	} else {
		// 非流式回应
		content := ""
		decoder := eventsource.NewDecoder(resp.Body)
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
		apiResp := gjson.New(ApiRespStr)
		apiResp.Set("choices.0.message.content", content)
		id := config.GenerateID(29)
		apiResp.Set("id", id)
		apiResp.Set("created", time.Now().Unix())
		if req.Model == "gpt-4" {
			apiResp.Set("model", "gpt-4")
		}
		r.Response.WriteJson(apiResp)
	}

}
