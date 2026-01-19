package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"os/signal"
	"time"

	"anayra-c6b9.net/mashupstudio/internal/middleware"
	"anayra-c6b9.net/mashupstudio/internal/rooms"
	"anayra-c6b9.net/mashupstudio/internal/handlers"
	"anayra-c6b9.net/mashupstudio/internal/websocket"

)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	roomHandler := handlers.NewRoomHandler(app.rooms)

	mux.HandleFunc("/api/rooms/create", roomHandler.CreateRoom)
	mux.HandleFunc("/api/rooms/join", roomHandler.JoinRoom)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	wsHandler := websocket.NewRoomWS(app.rooms)
	mux.HandleFunc("/ws/room", wsHandler.Handle)


	// Serve React build (unchanged)
	fileServer := http.FileServer(http.Dir("./ui/build"))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		path := "./ui/build" + r.URL.Path
		if _, err := os.Stat(path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, "./ui/build/index.html")
	}))

	return middleware.LogRequest(app.infoLog, mux)
}


type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	rooms    *rooms.Manager
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		infoLog:  infoLog,
		errorLog: errorLog,
		rooms:	rooms.NewManager(),
	}

	srv := &http.Server{
		Addr:    ":5500",
		Handler: app.routes(),
	}

	go func() {
		app.infoLog.Println("Starting Mashup Studio server on :5000")
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			app.errorLog.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	app.infoLog.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		app.errorLog.Fatal("Server forced to shutdown:", err)
	}

	app.infoLog.Println("Server exited cleanly")

}


