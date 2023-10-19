package apireq

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Req struct {
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	Stream    bool      `json:"stream"`
	PluginIds []string  `json:"plugin_ids"`
}
