package apirespstream

// ApiRespStream represents the JSON structure
type ApiRespStreamStruct struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []ChoiceObj `json:"choices"`
}

// ChoiceObj represents the nested "choices" object in the JSON
type ChoiceObj struct {
	Delta        DeltaObj `json:"delta"`
	Index        int      `json:"index"`
	FinishReason *string  `json:"finish_reason"`
}

// DeltaObj represents the nested "delta" object in the JSON
type DeltaObj struct {
	Content string `json:"content"`
}
