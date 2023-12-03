package chat

import (
	"chat2api/config"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/google/uuid"
	"github.com/launchdarkly/eventsource"
)

var (
	Dalle3req = `
	{
		"action": "next",
		"messages": [
		  {
			"id": "aaa2577b-9d88-406a-b1fb-2ea23d8a08ad",
			"author": { "role": "user" },
			"content": { "content_type": "text", "parts": ["画一只猫"] },
			"metadata": {}
		  }
		],
		"parent_message_id": "aaa14d98-383c-498c-8ba4-2753bb5afbdd",
		"model": "gpt-4-gizmo",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": true,
		"conversation_mode": {
		  "gizmo": {
			"gizmo": {
			  "id": "g-2fkFE8rbu",
			  "organization_id": "org-OROoM5KiDq6bcfid37dQx4z4",
			  "short_url": "g-2fkFE8rbu-dall-e",
			  "author": {
				"user_id": "user-u7SVk5APwT622QC7DPe41GHJ",
				"display_name": "ChatGPT",
				"link_to": null,
				"selected_display": "name",
				"is_verified": true
			  },
			  "voice": { "id": "ember" },
			  "workspace_id": null,
			  "model": null,
			  "instructions": null,
			  "settings": null,
			  "display": {
				"name": "DALL·E",
				"description": "Let me turn your imagination into imagery",
				"welcome_message": "Hello",
				"prompt_starters": null,
				"profile_picture_url": "https://files.oaiusercontent.com/file-SxYQO0Fq1ZkPagkFtg67DRVb?se=2123-10-12T23%3A57%3A32Z&sp=r&sv=2021-08-06&sr=b&rscc=max-age%3D31536000%2C%20immutable&rscd=attachment%3B%20filename%3Dagent_3.webp&sig=pLlQh8oUktqQzhM09SDDxn5aakqFuM2FAPptuA0mbqc%3D",
				"categories": []
			  },
			  "share_recipient": "marketplace",
			  "updated_at": "2023-11-12T19:29:32.777742+00:00",
			  "last_interacted_at": "2023-11-24T09:27:13.855518+00:00",
			  "tags": ["public", "first_party"],
			  "version": null,
			  "live_version": null,
			  "training_disabled": null,
			  "allowed_sharing_recipients": null,
			  "review_info": null,
			  "appeal_info": null,
			  "vanity_metrics": null
			},
			"tools": [
			  {
				"id": "gzm_cnf_KuQKBEnzFPMwdKIWYnOoetjx~gzm_tool_P9ZWt7cmybLejZWkNxDTEpIj",
				"type": "dalle",
				"settings": null,
				"metadata": null
			  }
			],
			"files": [],
			"product_features": {
			  "attachments": {
				"type": "retrieval",
				"accepted_mime_types": [
				  "text/x-ruby",
				  "text/x-tex",
				  "application/msword",
				  "text/x-script.python",
				  "text/plain",
				  "application/json",
				  "application/pdf",
				  "text/x-csharp",
				  "text/x-typescript",
				  "text/x-java",
				  "text/x-c",
				  "text/html",
				  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				  "application/x-latext",
				  "text/javascript",
				  "text/markdown",
				  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
				  "text/x-php",
				  "text/x-c++",
				  "text/x-sh"
				],
				"image_mime_types": [
				  "image/gif",
				  "image/webp",
				  "image/jpeg",
				  "image/png"
				],
				"can_accept_all_mime_types": true
			  }
			}
		  },
		  "kind": "gizmo_interaction",
		  "gizmo_id": "g-2fkFE8rbu"
		},
		"force_paragen": false,
		"force_rate_limit": false
	  }
	  `
)

type Dalle3RespData struct {
	RevisedPrompt string `json:"revised_prompt"`
	Url           string `json:"url"`
}
type Dalle3Resp struct {
	Created int64            `json:"created"`
	Data    []Dalle3RespData `json:"data"`
}

func Dalle3(r *ghttp.Request) {
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
	prompt := r.Get("prompt").String()
	if prompt == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"detail": "prompt is empty",
		})
		return
	}
	reqJson := gjson.New(Dalle3req)
	reqJson.Set("messages.0.content.parts.0", prompt)
	reqJson.Set("messages.0.id", uuid.NewString())
	reqJson.Set("parent_message_id", uuid.NewString())

	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
		"authkey":       config.AUTHKEY,
	}).Post(ctx, config.APISERVER, reqJson)
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"detail": err.Error(),
		})
		return
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		r.Response.Status = resp.StatusCode
		r.Response.WriteJson(g.Map{
			"detail": resp.ReadAllString(),
		})
		return
	}
	decoder := eventsource.NewDecoder(resp.Body)
	defer decoder.Decode()
	dalle3resp := Dalle3Resp{}
	dalle3resp.Created = gtime.Now().Unix()
	for {
		event, err := decoder.Decode()
		if err != nil {
			break
		}
		text := event.Data()
		// g.Log().Debug(ctx, text)
		if text == "" {
			continue
		}
		if text == "[DONE]" {
			break
		}
		resJson := gjson.New(text)
		role := resJson.Get("message.author.role").String()
		// g.Log().Debug(ctx, "role: ", role)
		if role != "tool" {
			continue
		}
		content_type := resJson.Get("message.content.content_type").String()
		if content_type != "multimodal_text" {
			continue
		}
		// g.Log().Debug(ctx, "content_type: ", content_type)
		parts := resJson.GetJsons("message.content.parts")
		// g.Dump(parts)
		if len(parts) == 0 {
			continue
		}
		for _, part := range parts {
			// partJson := gjson.New(part)
			revised_prompt := part.Get("metadata.dalle.prompt").String()
			url := part.Get("asset_pointer").String()
			url, err := GetDownloadUrl(ctx, url, token)
			if err != nil {
				g.Log().Error(ctx, err)
				continue
			}
			dalle3resp.Data = append(dalle3resp.Data, Dalle3RespData{
				RevisedPrompt: revised_prompt,
				Url:           url,
			})
		}
	}
	if len(dalle3resp.Data) == 0 {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"detail": "no data",
		})
		return
	}
	r.Response.WriteJson(dalle3resp)

}

// https://demo.xyhelper.cn/backend-api/files/file-FLeoX7FluBQ1Ri5JHDOq7ZiN/download

// file-service://file-1YBBS7IUuJaD3Qg7ro03mQ56
func GetDownloadUrl(ctx g.Ctx, url string, token string) (download_url string, err error) {
	// 将url c //分割 获取file_id
	fileId := gstr.Split(url, "//")[1]

	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
		"authkey":       config.AUTHKEY,
	}).Get(ctx, config.APIHOST+"/backend-api/files/"+fileId+"/download")
	if err != nil {
		return
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		err = gerror.New(resp.ReadAllString())
		return
	}
	respJson := gjson.New(resp.ReadAllString())
	download_url = respJson.Get("download_url").String()
	return

}
