#!/bin/bash
curl https://ai.fakeopen.com/v1/chat/completions  \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer pk-this-is-a-real-free-pool-token-for-everyone" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}],
    "stream": true

  }'
