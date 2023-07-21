# CHAT2API

将OPENAI官网接口转换为API格式，以兼容针对API开发的应用。

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
