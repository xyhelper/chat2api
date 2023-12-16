# CHAT2API


## 项目简介

CHAT2API 是一个开源项目，旨在将OPENAI官网接口转换为API格式，以兼容针对API开发的应用

**请注意：这个代码库不经常维护。**

由于资源有限，我们不能保证及时回应问题或合并拉取请求。我们鼓励社区成员互相帮助和合作，但请理解我们可能无法立即处理问题或合并更改。

## 特点

- 将OPENAI官网接口转换为API格式

## 如何贡献

尽管这个代码库不经常维护，但我们仍然欢迎社区的贡献。如果您想为项目做出贡献，您可以遵循以下步骤：

1. 选择一个任务，或者创建一个新的任务，开始编写代码。
2. 提交拉取请求（Pull Request）并等待审核。

虽然我们不能保证及时处理，但我们仍然感谢每一个对项目做出贡献的人！

## 安装和使用

```bash
curl https://api.xyhelper.cn/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-api-xyhelper-cn-free-token-for-everyone-xyhelper" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

## 社区支持

如果您有任何问题、建议或反馈，请随时联系我们或者在 [GitHub Issues](https://github.com/xyhelper/chat2api/issues) 上提交一个新的问题。
或者联系微信客户，或telegram客服
| ![wx](./images/wx.jpg) | ![telegram](./images/telegram.jpg) |
| ---------------------- | ---------------------------------- |
| 企业微信               | telegram                           |


**请注意：问题的响应时间可能较长。**


## 鸣谢

感谢您对 CHAT2API 的兴趣和支持！

如果您想了解更多关于项目的信息，请访问我们的 [xyhelper](https://www.xyhelper.com.cn/) 。
