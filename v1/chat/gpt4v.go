package chat

import (
	"chat2api/apirespstream"
	"chat2api/config"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/launchdarkly/eventsource"
)

func Gpt4v(r *ghttp.Request) {
	ctx := r.Context()
	authkey := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	if authkey == "" {
		r.Response.Status = 401
		r.Response.WriteJson(gjson.New(ErrNoAuth))
		return
	}
	g.Log().Info(ctx, "authkey: ", authkey)
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
	g.Log().Debug(ctx, "token: ", token)
	message := r.Get("message").String()
	if message == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"detail": "message is empty",
		})
		return
	}
	stream := r.Get("stream").Bool()

	// 获取上传的文件
	files := r.GetUploadFiles("file")
	if len(files) == 0 {
		r.Response.Status = 400
		r.Response.WriteJsonExit(g.Map{
			"code":   0,
			"detail": "upload file is empty",
		})
	}
	// 检查 ./temp 目录是否存在 不在则创建
	if !gfile.Exists("./temp") {
		err := gfile.Mkdir("./temp")
		if err != nil {
			r.Response.Status = 400
			r.Response.WriteJsonExit(g.Map{
				"code":   0,
				"detail": "create temp dir failed",
			})
		}
	}
	filenames, err := files.Save("./temp", true)
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(g.Map{
			"code":   0,
			"detail": "upload file failed",
		})
	}
	// 删除临时文件
	defer func() {
		for _, filename := range filenames {
			gfile.Remove("./temp/" + filename)
		}
	}()

	var file_ids []string
	var download_urls []string
	var widths []int
	var heights []int
	var size_bytess []int64
	// 上传文件到azure
	for _, filename := range filenames {
		file_id, download_url, width, height, size_bytes, err := UploadAzure(ctx, "./temp/"+filename, token)
		if err != nil {
			g.Log().Error(ctx, err)
			r.Response.Status = 400
			r.Response.WriteJsonExit(g.Map{
				"code":   0,
				"detail": err.Error(),
			})
		}
		file_ids = append(file_ids, file_id)
		download_urls = append(download_urls, download_url)
		widths = append(widths, width)
		heights = append(heights, height)
		size_bytess = append(size_bytess, size_bytes)
	}
	// g.Dump(file_ids)
	// g.Dump(download_urls)
	ChatReq := gjson.New(ChatReqStr)
	for i, file_id := range file_ids {
		ChatReq.Set("messages.0.content.parts."+gconv.String(i)+".asset_pointer", "file-service://"+file_id)
		ChatReq.Set("messages.0.content.parts."+gconv.String(i)+".height", heights[i])
		ChatReq.Set("messages.0.content.parts."+gconv.String(i)+".width", widths[i])
		ChatReq.Set("messages.0.content.parts."+gconv.String(i)+".size_bytes", size_bytess[i])
	}
	// messages.0.content.content_type multimodal_text
	ChatReq.Set("messages.0.content.content_type", "multimodal_text")
	ChatReq.Set("messages.0.content.parts."+gconv.String(len(file_ids)), message)
	ChatReq.Set("messages.0.id", uuid.NewString())
	ChatReq.Set("parent_message_id", uuid.NewString())
	ChatReq.Set("model", "gpt-4")
	// ChatReq.Remove("plugin_ids")

	// ChatReq.Dump()
	// 请求openai
	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}).Post(ctx, config.APISERVER, ChatReq.MustToJson())
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
	if stream {
		// 流式返回
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
		defer decoder.Decode()

		id := config.GenerateID(29)
		for {
			event, err := decoder.Decode()
			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}
			text := event.Data()
			if text == "" {
				continue
			}
			if text == "[DONE]" {
				apiRespStrEnd := gstr.Replace(ApiRespStrStreamEnd, "apirespid", id)
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apicreated", gconv.String(time.Now().Unix()))
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apirespmodel", "gpt-4")
				rw.Write([]byte("data: " + apiRespStrEnd + "\n\n"))

				rw.Write([]byte("data: " + text + "\n\n"))
				flusher.Flush()
				break
			}
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

				apiRespStruct := &apirespstream.ApiRespStreamStruct{}
				gconv.Struct(apiResp, apiRespStruct)
				apiRespStruct.Model = "gpt-4"
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
		decoder.Decode()

		completionTokens := CountTokens(content)
		promptTokens := CountTokens(message)
		totalTokens := completionTokens + promptTokens

		apiResp := gjson.New(ApiRespStr)
		apiResp.Set("choices.0.message.content", content)
		id := config.GenerateID(29)
		apiResp.Set("id", id)
		apiResp.Set("created", time.Now().Unix())
		apiResp.Set("model", "gpt-4")
		apiResp.Set("usage.prompt_tokens", promptTokens)
		apiResp.Set("usage.completion_tokens", completionTokens)
		apiResp.Set("usage.total_tokens", totalTokens)
		r.Response.WriteJson(apiResp)
	}

}

func UploadAzure(ctx g.Ctx, filepath string, token string) (file_id string, download_url string, width int, height int, size_bytes int64, err error) {
	// 检测文件是否存在
	if !gfile.Exists(filepath) {
		err = gerror.New("read file fail")
		return
	}

	fileName := gfile.Basename(filepath)
	fileSize := gfile.Size(filepath)
	apihost := config.APIHOST

	// 获取上传地址 backend-api/files  POST
	res, err := g.Client().SetHeader("Authorization", "Bearer "+token).ContentJson().Post(ctx, apihost+"/backend-api/files", g.Map{
		"file_name": fileName,
		"file_size": fileSize,
		"use_case":  "multimodal",
	})
	if err != nil {
		return
	}
	defer res.Close()
	if res.StatusCode != 200 {
		res.RawDump()
		err = gerror.New("get upload_url fail:" + res.Status)
		return
	}
	//
	resJson := gjson.New(res.ReadAllString())
	// resJson.Dump()
	upload_url := resJson.Get("upload_url").String()
	file_id = resJson.Get("file_id").String()
	if upload_url == "" {
		err = gerror.New("get upload_url fail")
		return
	}
	// 获取图片宽高
	file, err := gfile.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()
	// 获取图片宽高
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return
	}
	width = img.Width
	height = img.Height
	size_bytes = fileSize

	// 以二进制流的方式上传文件 PUT
	filedata := gfile.GetBytes(filepath)

	resput, err := g.Client().SetHeaderMap(g.MapStrStr{
		"x-ms-blob-type": "BlockBlob",
		"x-ms-version":   "2020-04-08",
	}).Put(ctx, upload_url, filedata)
	if err != nil {
		return
	}
	defer resput.Close()
	// resput.RawDump()
	if resput.StatusCode != 201 {
		err = gerror.New("upload file fail")
		return
	}
	// 获取文件下载地址 backend-api/files/{file_id}/uploaded  POST
	resdown, err := g.Client().SetHeader("Authorization", "Bearer "+token).ContentJson().Post(ctx, apihost+"/backend-api/files/"+file_id+"/uploaded")
	if err != nil {
		return
	}
	defer resdown.Close()
	resdown.RawDump()
	download_url = gjson.New(resdown.ReadAllString()).Get("download_url").String()
	if download_url == "" {
		err = gerror.New("get download_url fail")
		return
	}

	return
}
