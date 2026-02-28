package main

import "github.com/blvckbill/redis-from-scratch/internal/server"

func main() {
	server.NewServer().Start()
}
