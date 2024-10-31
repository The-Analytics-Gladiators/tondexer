package core

type Config struct {
	ConsoleToken string `yaml:"console_token" env:"CONSOLE_TOKEN" env-default:""`
	DbHost       string `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort       uint   `yaml:"db_port" env:"DB_PORT" env-default:"9000"`
	DbUser       string `yaml:"db_user" env:"DB_USER" env-default:"default"`
	DbPassword   string `yaml:"db_password" env:"DB_PASSWORD" env-default:""`
	DbName       string `yaml:"db_name" env:"DB_NAME" env-default:"default"`
}
