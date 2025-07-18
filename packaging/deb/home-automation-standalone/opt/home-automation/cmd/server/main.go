package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/handlers"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	fmt.Printf("Starting home automation server on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
