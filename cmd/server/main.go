package main

import (
	"log"
	"net/http"

	"go-project-2/internal/handler"
	"go-project-2/internal/model"
)

func main() {
	// 仓储和处理器在启动时创建，并在整个进程生命周期内复用。
	store := model.NewUserStore()
	userHandler := handler.NewUserHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/users", userHandler)
	mux.Handle("/users/", userHandler)

	const address = ":8080"
	log.Printf("HTTP 服务已启动，监听地址 %s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
