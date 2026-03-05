package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	HH       HHConfig       `mapstructure:"hh"`
	MCP      MCPConfig      `mapstructure:"mcp"`
	Search   SearchConfig   `mapstructure:"search"`
	Google   GoogleConfig   `mapstructure:"google"`
}

type GoogleConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	SheetID         string `mapstructure:"sheet_id"`
}

type PostgresConfig struct {
	URL string `mapstructure:"url"`
}

type HHConfig struct {
	Token          string `mapstructure:"token"`
	RateLimitMS    int    `mapstructure:"rate_limit_ms"`
	DefaultPerPage int    `mapstructure:"default_per_page"`
	MaxPages       int    `mapstructure:"max_pages"`
}

type MCPConfig struct {
	PythonBin    string `mapstructure:"python_bin"`
	ResumeServer string `mapstructure:"resume_server"`
	SheetsServer string `mapstructure:"sheets_server"`
	GroqAPIKey   string `mapstructure:"groq_api_key"`
}

type SearchConfig struct {
	Text           string `mapstructure:"text"`
	Area           string `mapstructure:"area"`
	Experience     string `mapstructure:"experience"`
	WorkFormat     string `mapstructure:"work_format"`
	Salary         int    `mapstructure:"salary"`
	OnlyWithSalary bool   `mapstructure:"only_with_salary"`
}

// Load загружает конфиг. Приоритет (от низшего к высшему):
// 1. дефолтные значения
// 2. config.yaml
// 3. .env файл
// 4. переменные окружения
func Load(cfgPath string) (*Config, error) {
	// загружаем .env в os.Environ до инициализации Viper
	_ = godotenv.Load(".env")

	v := viper.New()

	setDefaults(v)

	// yaml конфиг
	v.SetConfigFile(cfgPath)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// переменные окружения (наивысший приоритет)
	v.AutomaticEnv()
	bindEnvs(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("postgres.url", "postgres://postgres:postgres@localhost:5432/mcpprep?sslmode=disable")

	v.SetDefault("hh.token", "")
	v.SetDefault("hh.rate_limit_ms", 300)
	v.SetDefault("hh.default_per_page", 20)
	v.SetDefault("hh.max_pages", 5)

	v.SetDefault("mcp.python_bin", "python3")
	v.SetDefault("mcp.resume_server", "server/resume_server.py")
	v.SetDefault("mcp.sheets_server", "server/sheets_server.py")

	v.SetDefault("google.credentials_file", "credentials.json")
	v.SetDefault("google.sheet_id", "")

	v.SetDefault("search.text", "Go developer")
	v.SetDefault("search.area", "1")
	v.SetDefault("search.experience", "")
	v.SetDefault("search.work_format", "")
	v.SetDefault("search.salary", 0)
	v.SetDefault("search.only_with_salary", false)
}

func bindEnvs(v *viper.Viper) {
	v.BindEnv("postgres.url", "DATABASE_URL")

	v.BindEnv("hh.token", "HH_TOKEN")
	v.BindEnv("hh.rate_limit_ms", "HH_RATE_LIMIT_MS")
	v.BindEnv("hh.max_pages", "HH_MAX_PAGES")

	v.BindEnv("mcp.python_bin", "PYTHON_BIN")
	v.BindEnv("mcp.resume_server", "RESUME_SERVER")
	v.BindEnv("mcp.sheets_server", "SHEETS_SERVER")
	v.BindEnv("mcp.groq_api_key", "GROQ_API_KEY")

	v.BindEnv("google.credentials_file", "GOOGLE_CREDENTIALS_FILE")
	v.BindEnv("google.sheet_id", "GOOGLE_SHEET_ID")

	v.BindEnv("search.text", "SEARCH_TEXT")
	v.BindEnv("search.area", "SEARCH_AREA")
	v.BindEnv("search.experience", "SEARCH_EXPERIENCE")
	v.BindEnv("search.work_format", "SEARCH_WORK_FORMAT")
	v.BindEnv("search.salary", "SEARCH_SALARY")
	v.BindEnv("search.only_with_salary", "SEARCH_ONLY_WITH_SALARY")
}
