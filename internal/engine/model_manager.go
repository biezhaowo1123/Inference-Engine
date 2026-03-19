package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"inference-engine/internal/config"
	"net/http"
)

// ModelProvider 模型提供者接口
type ModelProvider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	StreamChat(ctx context.Context, messages []Message) (<-chan string, error)
	GetName() string
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelManager 模型管理器
type ModelManager struct {
	providers map[string]ModelProvider
	defaultModel string
}

// NewModelManager 创建模型管理器
func NewModelManager(cfg config.ModelsConfig) *ModelManager {
	mm := &ModelManager{
		providers:    make(map[string]ModelProvider),
		defaultModel: cfg.Default,
	}

	// ==================== 国际主流 ====================
	// DeepSeek
	if cfg.DeepSeek.Enabled && cfg.DeepSeek.APIKey != "" {
		mm.Register("deepseek", NewOpenAICompatibleProvider("deepseek", cfg.DeepSeek))
	}

	// Claude
	if cfg.Claude.Enabled && cfg.Claude.APIKey != "" {
		mm.Register("claude", NewClaudeProvider(cfg.Claude))
	}

	// GPT
	if cfg.GPT.Enabled && cfg.GPT.APIKey != "" {
		mm.Register("gpt", NewOpenAICompatibleProvider("gpt", cfg.GPT))
	}

	// Gemini
	if cfg.Gemini.Enabled && cfg.Gemini.APIKey != "" {
		mm.Register("gemini", NewGeminiProvider(cfg.Gemini))
	}

	// Grok
	if cfg.Grok.Enabled && cfg.Grok.APIKey != "" {
		mm.Register("grok", NewOpenAICompatibleProvider("grok", cfg.Grok))
	}

	// Llama
	if cfg.Llama.Enabled && cfg.Llama.APIKey != "" {
		mm.Register("llama", NewOpenAICompatibleProvider("llama", cfg.Llama))
	}

	// Mistral
	if cfg.Mistral.Enabled && cfg.Mistral.APIKey != "" {
		mm.Register("mistral", NewOpenAICompatibleProvider("mistral", cfg.Mistral))
	}

	// ==================== 国内大厂 ====================
	// 通义千问
	if cfg.Qwen.Enabled && cfg.Qwen.APIKey != "" {
		mm.Register("qwen", NewOpenAICompatibleProvider("qwen", cfg.Qwen))
	}

	// 智谱GLM
	if cfg.Glm.Enabled && cfg.Glm.APIKey != "" {
		mm.Register("glm", NewGLMProvider(cfg.Glm))
	}

	// 文心一言
	if cfg.Wenxin.Enabled && cfg.Wenxin.APIKey != "" {
		mm.Register("wenxin", NewWenxinProvider(cfg.Wenxin))
	}

	// 混元
	if cfg.Hunyuan.Enabled && cfg.Hunyuan.APIKey != "" {
		mm.Register("hunyuan", NewOpenAICompatibleProvider("hunyuan", cfg.Hunyuan))
	}

	// 星火
	if cfg.Spark.Enabled && cfg.Spark.APIKey != "" {
		mm.Register("spark", NewOpenAICompatibleProvider("spark", cfg.Spark))
	}

	// 豆包
	if cfg.Doubao.Enabled && cfg.Doubao.APIKey != "" {
		mm.Register("doubao", NewOpenAICompatibleProvider("doubao", cfg.Doubao))
	}

	// ==================== 国内创业公司 ====================
	// Moonshot
	if cfg.Moonshot.Enabled && cfg.Moonshot.APIKey != "" {
		mm.Register("moonshot", NewOpenAICompatibleProvider("moonshot", cfg.Moonshot))
	}

	// 百川
	if cfg.Baichuan.Enabled && cfg.Baichuan.APIKey != "" {
		mm.Register("baichuan", NewOpenAICompatibleProvider("baichuan", cfg.Baichuan))
	}

	// 零一万物
	if cfg.Yi.Enabled && cfg.Yi.APIKey != "" {
		mm.Register("yi", NewOpenAICompatibleProvider("yi", cfg.Yi))
	}

	// MiniMax
	if cfg.Minimax.Enabled && cfg.Minimax.APIKey != "" {
		mm.Register("minimax", NewOpenAICompatibleProvider("minimax", cfg.Minimax))
	}

	// ==================== 其他平台 ====================
	// Perplexity
	if cfg.Perplexity.Enabled && cfg.Perplexity.APIKey != "" {
		mm.Register("perplexity", NewOpenAICompatibleProvider("perplexity", cfg.Perplexity))
	}

	// Cohere
	if cfg.Cohere.Enabled && cfg.Cohere.APIKey != "" {
		mm.Register("cohere", NewCohereProvider(cfg.Cohere))
	}

	// Together
	if cfg.Together.Enabled && cfg.Together.APIKey != "" {
		mm.Register("together", NewOpenAICompatibleProvider("together", cfg.Together))
	}

	// OpenRouter
	if cfg.OpenRouter.Enabled && cfg.OpenRouter.APIKey != "" {
		mm.Register("openrouter", NewOpenAICompatibleProvider("openrouter", cfg.OpenRouter))
	}

	// ==================== 本地模型 ====================
	// Ollama (不需要API Key)
	if cfg.Ollama.Enabled {
		mm.Register("ollama", NewOpenAICompatibleProvider("ollama", cfg.Ollama))
	}

	// vLLM
	if cfg.VLLM.Enabled {
		mm.Register("vllm", NewOpenAICompatibleProvider("vllm", cfg.VLLM))
	}

	// LocalAI
	if cfg.LocalAI.Enabled {
		mm.Register("localai", NewOpenAICompatibleProvider("localai", cfg.LocalAI))
	}

	return mm
}

// Register 注册模型提供者
func (mm *ModelManager) Register(name string, provider ModelProvider) {
	mm.providers[name] = provider
}

// GetProvider 获取模型提供者
func (mm *ModelManager) GetProvider(name string) (ModelProvider, error) {
	if name == "" {
		name = mm.defaultModel
	}
	
	provider, ok := mm.providers[name]
	if !ok {
		return nil, fmt.Errorf("模型 %s 未配置或未启用", name)
	}
	return provider, nil
}

// ListModels 列出可用模型
func (mm *ModelManager) ListModels() []string {
	models := make([]string, 0, len(mm.providers))
	for name := range mm.providers {
		models = append(models, name)
	}
	return models
}

// ==================== 通用OpenAI兼容Provider ====================

type OpenAICompatibleProvider struct {
	name    string
	apiKey  string
	baseURL string
	model   string
}

func NewOpenAICompatibleProvider(name string, cfg config.ModelConfig) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		name:    name,
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *OpenAICompatibleProvider) GetName() string {
	return p.name
}

func (p *OpenAICompatibleProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callOpenAICompatible(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *OpenAICompatibleProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamOpenAICompatible(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== DeepSeek Provider (兼容OpenAI) ====================

type DeepSeekProvider struct {
	*OpenAICompatibleProvider
}

func NewDeepSeekProvider(cfg config.ModelConfig) *DeepSeekProvider {
	return &DeepSeekProvider{
		OpenAICompatibleProvider: NewOpenAICompatibleProvider("deepseek", cfg),
	}
}

// ==================== GPT Provider (兼容OpenAI) ====================

type GPTProvider struct {
	*OpenAICompatibleProvider
}

func NewGPTProvider(cfg config.ModelConfig) *GPTProvider {
	return &GPTProvider{
		OpenAICompatibleProvider: NewOpenAICompatibleProvider("gpt", cfg),
	}
}

// ==================== Claude Provider ====================

type ClaudeProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewClaudeProvider(cfg config.ModelConfig) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *ClaudeProvider) GetName() string {
	return "claude"
}

func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callClaudeAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *ClaudeProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamClaudeAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== Gemini Provider ====================

type GeminiProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewGeminiProvider(cfg config.ModelConfig) *GeminiProvider {
	return &GeminiProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *GeminiProvider) GetName() string {
	return "gemini"
}

func (p *GeminiProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callGeminiAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *GeminiProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamGeminiAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== GLM Provider ====================

type GLMProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewGLMProvider(cfg config.ModelConfig) *GLMProvider {
	return &GLMProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *GLMProvider) GetName() string {
	return "glm"
}

func (p *GLMProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callOpenAICompatible(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *GLMProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamOpenAICompatible(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== Wenxin Provider (百度文心) ====================

type WenxinProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewWenxinProvider(cfg config.ModelConfig) *WenxinProvider {
	return &WenxinProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *WenxinProvider) GetName() string {
	return "wenxin"
}

func (p *WenxinProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callWenxinAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *WenxinProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamWenxinAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== Cohere Provider ====================

type CohereProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewCohereProvider(cfg config.ModelConfig) *CohereProvider {
	return &CohereProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

func (p *CohereProvider) GetName() string {
	return "cohere"
}

func (p *CohereProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return callCohereAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

func (p *CohereProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	return streamCohereAPI(ctx, p.baseURL, p.apiKey, p.model, messages)
}

// ==================== HTTP调用实现 ====================

func callOpenAICompatible(ctx context.Context, baseURL, apiKey, model string, messages []Message) (string, error) {
	// 构建请求体
	reqBody := map[string]interface{}{
		"model": model,
		"messages": func() []map[string]string {
			result := make([]map[string]string, len(messages))
			for i, m := range messages {
				result[i] = map[string]string{
					"role":    m.Role,
					"content": m.Content,
				}
			}
			return result
		}(),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/chat/completions", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("API错误: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("无返回结果")
	}

	return result.Choices[0].Message.Content, nil
}

func streamOpenAICompatible(ctx context.Context, baseURL, apiKey, model string, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		// 流式返回
		ch <- "推理结果"
	}()
	return ch, nil
}

func callClaudeAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (string, error) {
	// Claude API调用
	return "Claude推理结果...", nil
}

func streamClaudeAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "Claude推理结果"
	}()
	return ch, nil
}

func callGeminiAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (string, error) {
	// Gemini API调用
	// 使用Generative Language API格式
	return "Gemini推理结果...", nil
}

func streamGeminiAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "Gemini推理结果"
	}()
	return ch, nil
}

func callWenxinAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (string, error) {
	// 百度文心API调用
	return "文心一言推理结果...", nil
}

func streamWenxinAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "文心一言推理结果"
	}()
	return ch, nil
}

func callCohereAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (string, error) {
	// Cohere API调用
	return "Cohere推理结果...", nil
}

func streamCohereAPI(ctx context.Context, baseURL, apiKey, model string, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "Cohere推理结果"
	}()
	return ch, nil
}
