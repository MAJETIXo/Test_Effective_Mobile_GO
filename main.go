package main

import (
	"fmt"
	"log"
	"net/http"
	"server/migrations"
	"server/routers"
)

func main() {
  migrations.MigrateDB()
  r := routers.InitRoutes()
  fmt.Println("Server is running on port 8000...")
  log.Fatal(http.ListenAndServe(":8000", r))
}