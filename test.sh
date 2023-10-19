curl http://127.0.0.1:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-api-xyhelper-cn-free-token-for-everyone-xyhelper" \
  -d '{
    "model": "gpt-4-32k",
    "messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "计算一下圆周率"}],
    "stream": true,
    "plugin_ids": ["plugin-d1d6eb04-3375-40aa-940a-c2fc57ce0f51"]
  }'
