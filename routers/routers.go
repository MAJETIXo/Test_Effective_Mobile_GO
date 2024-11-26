package routers

import (
	"net/http"
	"server/handlers"

	_ "server/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/info", handlers.GetSongInfo).Methods("GET")
	r.HandleFunc("/music", handlers.PostMusic).Methods("POST")
	r.HandleFunc("/music/{id:[0-9]+}", handlers.DeleteMusic).Methods("DELETE")
	r.HandleFunc("/music/{id:[0-9]+}", handlers.UpdateMusic).Methods("PUT")
	r.HandleFunc("/music/{id:[0-9]+}/text", handlers.GetSongText).Methods("GET")
	r.HandleFunc("/songs", handlers.GetGroupWithSongs).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, ".docs/swagger.json")
	})

	return r
}