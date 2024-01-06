package openai

type ChatCompletionResponse struct {
    ID               string     `json:"id"`
    Object           string     `json:"object"`
    Created          int64      `json:"created"`
    Model            string     `json:"model"`
    SystemFingerprint string    `json:"system_fingerprint"`
    Choices          []Choice   `json:"choices"`
    Usage            Usage      `json:"usage"`
}

type Choice struct {
    Index          int         `json:"index"`
    Message        Message     `json:"message"`
    LogProbs       interface{} `json:"logprobs"` // null in JSON; use interface{} in Go
    FinishReason   string      `json:"finish_reason"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Usage struct {
    PromptTokens      int `json:"prompt_tokens"`
    CompletionTokens  int `json:"completion_tokens"`
    TotalTokens       int `json:"total_tokens"`
}

