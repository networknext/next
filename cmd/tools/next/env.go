package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

type Environment struct {
	Hostname string `json:"hostname"`
}

func (e *Environment) String() string {
	return "Hostname: " + e.Hostname
}

func (e *Environment) Exists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to read environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	if _, err := os.Stat(envFilePath); err != nil {
		return false
	}

	return true
}

func (e *Environment) Read() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to read environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Open(envFilePath)
	if err != nil {
		log.Fatal("failed to read environment", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(e); err != nil {
		log.Fatal("failed to read environment", err)
	}
}

func (e *Environment) Write() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to write environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Create(envFilePath)
	if err != nil {
		log.Fatal("failed to write environment", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		log.Fatal("failed to write environment", err)
	}
}

func (e *Environment) Clean() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to clean environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	err = os.RemoveAll(envFilePath)
	if err != nil {
		log.Fatal("failed to clean environment", err)
	}
}
