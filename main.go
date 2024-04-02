package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type RepositoryConfig struct {
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Apps []App  `json:"apps"`
}

type App struct {
	AppName      string        `json:"appName"`
	Environments []Environment `json:"environments"`
}

type Environment struct {
	Name             string           `json:"name"`
	AwsAccountId     string           `json:"awsAccountId"`
	IamRoles         []IamRole        `json:"iamRoles"`
	TerraformBackend TerraformBackend `json:"terraformBackend"`
}

type IamRole struct {
	RoleName  string `json:"roleName"`
	PolicyArn string `json:"policyArn"`
}

type TerraformBackend struct {
	S3 S3Backend `json:"s3"`
}

type S3Backend struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Region string `json:"region"`
}

func loadConfigData(filePath string) (*RepositoryConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config RepositoryConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getConfigHandler(config *RepositoryConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repo")
		appName := r.PathValue("app")
		envName := r.PathValue("environment")

		if r.RequestURI == "/configs" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config)
			return
		}

		for _, repo := range config.Repositories {
			if repo.Name != repoName {
				continue
			}
			for _, app := range repo.Apps {
				if app.AppName != appName {
					continue
				}
				for _, env := range app.Environments {
					if env.Name != envName {
						continue
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(env)
					return
				}
			}
		}
		http.Error(w, "Configuration not found", http.StatusNotFound)
	}
}

func main() {
	fmt.Println("Starting the server...")

	config, err := loadConfigData("config.json") // TODO: Store data in dynamoDB
	if err != nil {
		fmt.Println("Error loading configuration data:", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /configs", getConfigHandler(config))
	mux.HandleFunc("GET /configs/{repo}", getConfigHandler(config))
	mux.HandleFunc("GET /configs/{repo}/{app}/{environment}", getConfigHandler(config))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Error starting the server:", err)
	}
}
