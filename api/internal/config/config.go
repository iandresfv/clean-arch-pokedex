// Package config provides environment-based configuration loading.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server   Server
	Database Database
}

type Server struct {
	Host string
	Port int
}

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (*Config, error) {
	srv := Server{
		Host: getEnv("SERVER_HOST", "localhost"),
		Port: getEnvInt("SERVER_PORT", 8080),
	}
	db := Database{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "pokedex"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
	config := &Config{
		Server:   srv,
		Database: db,
	}
	return config, nil
}

func (s Server) Addr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (d Database) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	strValue, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		log.Printf("invalid integer for %s: %v", key, err)
		return fallback
	}
	return intValue
}
