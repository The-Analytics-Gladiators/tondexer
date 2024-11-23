package main

type Config struct {
	ConsoleToken      string   `yaml:"console_token" env:"CONSOLE_TOKEN" env-default:""`
	DbHost            string   `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort            uint     `yaml:"db_port" env:"DB_PORT" env-default:"9000"`
	DbUser            string   `yaml:"db_user" env:"DB_USER" env-default:"default"`
	DbPassword        string   `yaml:"db_password" env:"DB_PASSWORD" env-default:""`
	DbName            string   `yaml:"db_name" env:"DB_NAME" env-default:"default"`
	StonfiV1Addresses []string `yaml:"stonfiv1_addresses" env:"STONFIV1_ADDRESSES" env-default:""`
	StonfiV2Addresses []string `yaml:"stonfiv2_addresses" env:"STONFIV2_ADDRESSES" env-default:""`
	DedustAddresses   []string `yaml:"dedust_addresses" env:"DEDUST_ADDRESSES" env-default:""`
}
