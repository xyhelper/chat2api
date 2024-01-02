### 如何使用 以及 config.yaml参数作用


## 配置config文件

默认config在config/config.yaml中

```config/config.yaml
# PORT: 8089
APISERVER: https://freechat.xyhelper.cn/backend-api/conversation
PASSMODE: true
MAXTIME: 60
NOPLUGINS: true
KEEPHISTORY: true


sk-api-xyhelper-cn-free-token-for-everyone-xyhelper: "xyhelper.cn"

```

其中在不提供config.yaml时默认值为

```config/config.go
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

```
默认值是在有容器chatproxy提供代理服务的前提下。如果自己部署的话不进行修改/部署chatproxy，将无法使用。
- PORT:提供服务的端口
- APISERVER:对应的网页*聊天*提供服务的后端api网址，如果服务器处在openai服务范围内可以直接使用 https://chat.openai.com/backend-api/conversation
- APIHOST:就是去掉尾巴的APISERVER，为DALLE.3或者上传文件时使用，若服务器可以直接访问可填写 https://chat.openai.com  如果不需要使用除了聊天以外功能则可不填。
- PASSMODE:BOOL类型变量，决定是否直接把bearer参数传出而非通过SK2TOKEN(config/config.go)中转为token
- MAXTIME:最大等待响应时间
- KEEPHISTORY:chatgpt网页是否保存历史记录设置为true时，前往chat.openai.com登录对应账号会查看到之前对话。**对话如果选择保存会遵循相关政策，目前应该时允许使用数据作为训练集，详细查看openai官网**


所以一个可以直接访问chat.openai.com的服务器的配置文件可能如下
```config
# PORT: 8089
APISERVER: https://chat.openai.com/backend-api/conversation
APIHOST: https://chat.openai.com
PASSMODE: true
MAXTIME: 60
NOPLUGINS: true
KEEPHISTORY: false
```

在这个情况下传入的请求原本openai api key应该变为accesstoken。


## 获取accesstoken
- EDGE/CHROME浏览器 登录chat.openai.com时输入完账号和密码。按F12-Network会查看到session
    其中会有如下格式内容
    ```
    {
        "user": {
            "id": "user-id",
            "name": "yourname",
            "email": "someemail",
            "image": "xxxxxx",
            "picture": "zzzz",
            "idp": "auth0",
            "iat": xxxx,
            "mfa": false,
            "groups": [],
            "intercom_hash": "xxxx"
        },
        "expires": "2024-04-01T09:33:07.368Z",
        "accessToken": "some access token",
        "authProvider": "auth0"
    }
    ```
    其中acesstoken之后便是需要的
- PandoraNext自建站可以获取accesstoken，详情参考[pandora-next/deploy](https://github.com/pandora-next/deploy)