package utils

import "time"

var (
	ValidScopes = [...]string{"read:a", "read:b", "write:a", "write:b"}
	Location    = time.Now().Location()
)

const (
	API_VERSION = "1.0.0"
)
