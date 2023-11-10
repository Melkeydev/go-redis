package main

import (
	"go-auth-testing/internal/auth"
	"go-auth-testing/internal/server"
)

func main() {

	auth.NewAuth()
	server.InitRedis()
	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic("cannot start server")
	}
}
