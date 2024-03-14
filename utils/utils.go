package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occurred. Err: %s", err)
	}
}

func Contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false

}

func Remove(haystack []string, needle string) []string {
	for i, item := range haystack {
		if item == needle {
			return append(haystack[:i], haystack[i+1:]...)
		}
	}
	return haystack
}
