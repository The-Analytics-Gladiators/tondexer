package main

type Config struct {
	DbHost     string `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort     uint   `yaml:"db_port" env:"DB_PORT" env-default:"9000"`
	DbUser     string `yaml:"db_user" env:"DB_USER" env-default:"default"`
	DbPassword string `yaml:"db_password" env:"DB_PASSWORD" env-default:""`
	DbName     string `yaml:"db_name" env:"DB_NAME" env-default:"default"`
}
