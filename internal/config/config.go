package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Models   ModelsConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ModelsConfig struct {
	// ==================== 国际主流 ====================
	DeepSeek  ModelConfig
	Claude    ModelConfig
	GPT       ModelConfig
	Gemini    ModelConfig
	Grok      ModelConfig
	Llama     ModelConfig
	Mistral   ModelConfig
	
	// ==================== 国内大厂 ====================
	Qwen      ModelConfig // 阿里云通义千问
	Glm       ModelConfig // 智谱AI
	Wenxin    ModelConfig // 百度文心一言
	Hunyuan   ModelConfig // 腾讯混元
	Spark     ModelConfig // 讯飞星火
	Doubao    ModelConfig // 字节豆包
	
	// ==================== 国内创业公司 ====================
	Moonshot  ModelConfig // 月之暗面
	Baichuan  ModelConfig // 百川智能
	Yi        ModelConfig // 零一万物
	Minimax   ModelConfig // MiniMax
	
	// ==================== 其他平台 ====================
	Perplexity ModelConfig // Perplexity
	Cohere    ModelConfig  // Cohere
	Together  ModelConfig  // Together AI
	OpenRouter ModelConfig // OpenRouter路由
	
	// ==================== 本地模型 ====================
	Ollama    ModelConfig
	VLLM      ModelConfig // vLLM本地部署
	LocalAI   ModelConfig // LocalAI
	
	Default   string
}

type ModelConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Enabled bool
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "inference.db"),
		},
		Models: ModelsConfig{
			Default: getEnv("DEFAULT_MODEL", "deepseek"),
			
			// ==================== 国际主流 ====================
			DeepSeek: ModelConfig{
				APIKey:  getEnv("DEEPSEEK_API_KEY", ""),
				BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
				Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
				Enabled: getEnvBool("DEEPSEEK_ENABLED", false),
			},
			
			Claude: ModelConfig{
				APIKey:  getEnv("CLAUDE_API_KEY", ""),
				BaseURL: getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"),
				Model:   getEnv("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
				Enabled: getEnvBool("CLAUDE_ENABLED", false),
			},
			
			GPT: ModelConfig{
				APIKey:  getEnv("GPT_API_KEY", ""),
				BaseURL: getEnv("GPT_BASE_URL", "https://api.openai.com/v1"),
				Model:   getEnv("GPT_MODEL", "gpt-4o"),
				Enabled: getEnvBool("GPT_ENABLED", false),
			},
			
			Gemini: ModelConfig{
				APIKey:  getEnv("GEMINI_API_KEY", ""),
				BaseURL: getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
				Model:   getEnv("GEMINI_MODEL", "gemini-2.0-flash-exp"),
				Enabled: getEnvBool("GEMINI_ENABLED", false),
			},
			
			Grok: ModelConfig{
				APIKey:  getEnv("GROK_API_KEY", ""),
				BaseURL: getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),
				Model:   getEnv("GROK_MODEL", "grok-beta"),
				Enabled: getEnvBool("GROK_ENABLED", false),
			},
			
			Llama: ModelConfig{
				APIKey:  getEnv("LLAMA_API_KEY", ""),
				BaseURL: getEnv("LLAMA_BASE_URL", "https://api.llama-api.com/v1"),
				Model:   getEnv("LLAMA_MODEL", "llama-3.3-70b"),
				Enabled: getEnvBool("LLAMA_ENABLED", false),
			},
			
			Mistral: ModelConfig{
				APIKey:  getEnv("MISTRAL_API_KEY", ""),
				BaseURL: getEnv("MISTRAL_BASE_URL", "https://api.mistral.ai/v1"),
				Model:   getEnv("MISTRAL_MODEL", "mistral-large-latest"),
				Enabled: getEnvBool("MISTRAL_ENABLED", false),
			},
			
			// ==================== 国内大厂 ====================
			Qwen: ModelConfig{
				APIKey:  getEnv("QWEN_API_KEY", ""),
				BaseURL: getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
				Model:   getEnv("QWEN_MODEL", "qwen-plus"),
				Enabled: getEnvBool("QWEN_ENABLED", false),
			},
			
			Glm: ModelConfig{
				APIKey:  getEnv("GLM_API_KEY", ""),
				BaseURL: getEnv("GLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
				Model:   getEnv("GLM_MODEL", "glm-4"),
				Enabled: getEnvBool("GLM_ENABLED", false),
			},
			
			Wenxin: ModelConfig{
				APIKey:  getEnv("WENXIN_API_KEY", ""),
				BaseURL: getEnv("WENXIN_BASE_URL", "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop"),
				Model:   getEnv("WENXIN_MODEL", "ernie-4.0-8k"),
				Enabled: getEnvBool("WENXIN_ENABLED", false),
			},
			
			Hunyuan: ModelConfig{
				APIKey:  getEnv("HUNYUAN_API_KEY", ""),
				BaseURL: getEnv("HUNYUAN_BASE_URL", "https://api.hunyuan.cloud.tencent.com/v1"),
				Model:   getEnv("HUNYUAN_MODEL", "hunyuan-lite"),
				Enabled: getEnvBool("HUNYUAN_ENABLED", false),
			},
			
			Spark: ModelConfig{
				APIKey:  getEnv("SPARK_API_KEY", ""),
				BaseURL: getEnv("SPARK_BASE_URL", "https://spark-api-open.xf-yun.com/v1"),
				Model:   getEnv("SPARK_MODEL", "generalv3.5"),
				Enabled: getEnvBool("SPARK_ENABLED", false),
			},
			
			Doubao: ModelConfig{
				APIKey:  getEnv("DOUBAO_API_KEY", ""),
				BaseURL: getEnv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
				Model:   getEnv("DOUBAO_MODEL", "doubao-pro-32k"),
				Enabled: getEnvBool("DOUBAO_ENABLED", false),
			},
			
			// ==================== 国内创业公司 ====================
			Moonshot: ModelConfig{
				APIKey:  getEnv("MOONSHOT_API_KEY", ""),
				BaseURL: getEnv("MOONSHOT_BASE_URL", "https://api.moonshot.cn/v1"),
				Model:   getEnv("MOONSHOT_MODEL", "moonshot-v1-8k"),
				Enabled: getEnvBool("MOONSHOT_ENABLED", false),
			},
			
			Baichuan: ModelConfig{
				APIKey:  getEnv("BAICHUAN_API_KEY", ""),
				BaseURL: getEnv("BAICHUAN_BASE_URL", "https://api.baichuan-ai.com/v1"),
				Model:   getEnv("BAICHUAN_MODEL", "Baichuan4"),
				Enabled: getEnvBool("BAICHUAN_ENABLED", false),
			},
			
			Yi: ModelConfig{
				APIKey:  getEnv("YI_API_KEY", ""),
				BaseURL: getEnv("YI_BASE_URL", "https://api.lingyiwanwu.com/v1"),
				Model:   getEnv("YI_MODEL", "yi-large"),
				Enabled: getEnvBool("YI_ENABLED", false),
			},
			
			Minimax: ModelConfig{
				APIKey:  getEnv("MINIMAX_API_KEY", ""),
				BaseURL: getEnv("MINIMAX_BASE_URL", "https://api.minimax.chat/v1"),
				Model:   getEnv("MINIMAX_MODEL", "abab6.5-chat"),
				Enabled: getEnvBool("MINIMAX_ENABLED", false),
			},
			
			// ==================== 其他平台 ====================
			Perplexity: ModelConfig{
				APIKey:  getEnv("PERPLEXITY_API_KEY", ""),
				BaseURL: getEnv("PERPLEXITY_BASE_URL", "https://api.perplexity.ai"),
				Model:   getEnv("PERPLEXITY_MODEL", "llama-3.1-sonar-large-128k-online"),
				Enabled: getEnvBool("PERPLEXITY_ENABLED", false),
			},
			
			Cohere: ModelConfig{
				APIKey:  getEnv("COHERE_API_KEY", ""),
				BaseURL: getEnv("COHERE_BASE_URL", "https://api.cohere.ai/v1"),
				Model:   getEnv("COHERE_MODEL", "command-r-plus"),
				Enabled: getEnvBool("COHERE_ENABLED", false),
			},
			
			Together: ModelConfig{
				APIKey:  getEnv("TOGETHER_API_KEY", ""),
				BaseURL: getEnv("TOGETHER_BASE_URL", "https://api.together.xyz/v1"),
				Model:   getEnv("TOGETHER_MODEL", "meta-llama/Llama-3-70b-chat-hf"),
				Enabled: getEnvBool("TOGETHER_ENABLED", false),
			},
			
			OpenRouter: ModelConfig{
				APIKey:  getEnv("OPENROUTER_API_KEY", ""),
				BaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
				Model:   getEnv("OPENROUTER_MODEL", "anthropic/claude-3.5-sonnet"),
				Enabled: getEnvBool("OPENROUTER_ENABLED", false),
			},
			
			// ==================== 本地模型 ====================
			Ollama: ModelConfig{
				APIKey:  getEnv("OLLAMA_API_KEY", ""),
				BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434/v1"),
				Model:   getEnv("OLLAMA_MODEL", "qwen2.5:7b"),
				Enabled: getEnvBool("OLLAMA_ENABLED", false),
			},
			
			VLLM: ModelConfig{
				APIKey:  getEnv("VLLM_API_KEY", ""),
				BaseURL: getEnv("VLLM_BASE_URL", "http://localhost:8000/v1"),
				Model:   getEnv("VLLM_MODEL", ""),
				Enabled: getEnvBool("VLLM_ENABLED", false),
			},
			
			LocalAI: ModelConfig{
				APIKey:  getEnv("LOCALAI_API_KEY", ""),
				BaseURL: getEnv("LOCALAI_BASE_URL", "http://localhost:8080/v1"),
				Model:   getEnv("LOCALAI_MODEL", ""),
				Enabled: getEnvBool("LOCALAI_ENABLED", false),
			},
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, _ := strconv.ParseBool(value)
		return b
	}
	return defaultValue
}
