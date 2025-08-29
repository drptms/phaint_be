package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type FirebaseConfig struct {
	DatabaseURL     string `yaml:"database_url"`
	CredentialsFile string `yaml:"credential_path"`
	WebApiKey       string `yaml:"web_api_key"`
}

type Config struct {
	Firebase FirebaseConfig `yaml:"firebase"`
}

var config Config

// Load the configuration variables from the secrets
func loadConfig() *Config {
	home, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error finding home directory: %v", err)
	}
	file, err := os.Open(home + "\\config\\secrets\\config.yaml")
	if err != nil {
		log.Fatalf("Error opening configuration file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding configuration file: %v", err)
	}

	return &config
}

// FirebaseDBUrl Return the Url of the Firebase database
func FirebaseDBUrl() string {
	loadConfig()
	return config.Firebase.DatabaseURL
}

// FirebaseCredentialsPath Return the path of the file containing the firebase credentials
func FirebaseCredentialsPath() string {
	loadConfig()
	home, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error finding home directory: %v", err)
	}
	location := home + config.Firebase.CredentialsFile
	if _, err := os.Stat(location); os.IsNotExist(err) {
		log.Fatalf("Firebase credentials file not found at %s", location)
	}
	return location
}

// FirebaseWebAPIKey Return the Firebase Web API key, useful for specific post request
func FirebaseWebAPIKey() string {
	loadConfig()
	return config.Firebase.WebApiKey
}
