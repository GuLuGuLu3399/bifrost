// common/src/config.rs

use serde::Deserialize;

// 给外部 config crate 起别名，避免和 crate::config 路径冲突
use config as config_crate;
// 应用配置结构体
#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct AppConfig {
    pub server: Option<ServerConfig>,
    pub features: Option<FeaturesConfig>,
    pub log: Option<LogConfig>,
    pub redis: Option<RedisConfig>,
    pub database: Option<DatabaseConfig>,
    pub nats: Option<NatsConfig>,
    pub otel: Option<OtelConfig>,
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct FeaturesConfig {
    pub enable_nats_worker: Option<bool>,
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct ServerConfig {
    pub addr: String,
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct LogConfig {
    pub level: Option<String>,
    pub format: Option<String>, // json / console
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct RedisConfig {
    pub dsn: String, // redis://:pass@127.0.0.1:6379/0
    pub max_connections: Option<u32>,
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct DatabaseConfig {
    pub dsn: String, // postgres://user:pass@localhost:5432/db
    pub max_connections: Option<u32>,
}

#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct NatsConfig {
    pub url: String,
    pub filter_subject: Option<String>,
    pub stream_name: Option<String>,
    pub consumer_name: Option<String>,
}

// [新增] OpenTelemetry 配置
#[derive(Debug, Deserialize, Clone, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub struct OtelConfig {
    pub enabled: bool,
    pub endpoint: String, // e.g., "http://localhost:4317"
}
/// Viper\-like 加载器：默认值 \< 文件 \< 环境变量
pub struct ConfigLoader {
    file: Option<std::path::PathBuf>,
    env_prefix: String,
    env_separator: String,
    dotenv: bool,
}

impl Default for ConfigLoader {
    fn default() -> Self {
        Self {
            file: None,
            env_prefix: "APP".to_string(),
            env_separator: "__".to_string(),
            dotenv: true,
        }
    }
}

impl ConfigLoader {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_file(mut self, path: impl Into<std::path::PathBuf>) -> Self {
        self.file = Some(path.into());
        self
    }

    pub fn with_env_prefix(mut self, prefix: impl Into<String>) -> Self {
        self.env_prefix = prefix.into();
        self
    }

    pub fn with_env_separator(mut self, sep: impl Into<String>) -> Self {
        self.env_separator = sep.into();
        self
    }

    pub fn with_dotenv(mut self, enabled: bool) -> Self {
        self.dotenv = enabled;
        self
    }

    pub fn load(self) -> Result<AppConfig, ConfigError> {
        if self.dotenv {
            let _ = dotenvy::dotenv();
        }

        let mut builder = config_crate::Config::builder();

        builder = builder
            .set_default("log.format", "console")
            .map_err(ConfigError::Config)?;

        if let Some(path) = self.file {
            let source = config_crate::File::from(path).required(false);
            builder = builder.add_source(source);
        }

        let env = config_crate::Environment::with_prefix(&self.env_prefix)
            .separator(&self.env_separator)
            .try_parsing(true);
        builder = builder.add_source(env);

        let cfg = builder.build().map_err(ConfigError::Config)?;
        cfg.try_deserialize::<AppConfig>()
            .map_err(ConfigError::Deserialize)
    }
}

#[derive(Debug)]
pub enum ConfigError {
    Config(config_crate::ConfigError),
    Deserialize(config_crate::ConfigError),
}

impl std::fmt::Display for ConfigError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Config(e) => write!(f, "config build error: {e}"),
            Self::Deserialize(e) => write!(f, "config deserialize error: {e}"),
        }
    }
}

impl std::error::Error for ConfigError {}
