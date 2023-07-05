```bash
curl https://openai-api-proxy.xyhelper.cn/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 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" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}]
  }'
```