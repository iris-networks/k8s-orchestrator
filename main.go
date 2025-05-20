package main

import (
	"os"

	"github.com/k8sgo/cmd/server"
)

func main() {
	os.Exit(server.Run())
}