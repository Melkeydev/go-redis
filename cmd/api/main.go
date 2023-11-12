package main

import (
	"go-auth-testing/internal/auth"
	"go-auth-testing/internal/server"
	"log"
)

func main() {

	auth.NewAuth()
	server.InitRedis()
	server := server.NewServer()

	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("Error closing the server: %v", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		panic("cannot start server")
	}
}
