package main

import "time"

type Config struct {
	System SystemConfig `yaml:"system"`
	Data   DataConfig   `yaml:"data"`
	Log    LogConfig    `yaml:"log"`
}

type SystemConfig struct {
	App string `yaml:"app"`
}

type DataConfig struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

type DatabaseConfig struct {
	Movie MovieDatabaseConfig `yaml:"movie"`
}

type MovieDatabaseConfig struct {
	DSNDes          string        `yaml:"dsn_des"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time"`
	Name            string        `yaml:"name"`
	LogErr          bool          `yaml:"log_err"`
	LogSlow         time.Duration `yaml:"log_slow"`
	Metrics         bool          `yaml:"metrics"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type LogConfig struct {
	LogToStdout  bool          `yaml:"log_to_stdout"`
	LogToFile    bool          `yaml:"log_to_file"`
	Path         string        `yaml:"path"`
	FileName     string        `yaml:"file_name"`
	RotationTime time.Duration `yaml:"rotation_time"`
	FileAge      time.Duration `yaml:"file_age"`
	Level        string        `yaml:"level"`
	P            string        `yaml:"p"`
	B            string        `yaml:"b"`
}
