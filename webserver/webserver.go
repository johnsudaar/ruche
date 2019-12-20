package webserver

import (
	"context"
	"fmt"
	"net/http"

	handlers "github.com/Scalingo/go-handlers"
	muxhandlers "github.com/gorilla/handlers"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/ruche/config"
)

func Start(ctx context.Context) {
	log := logger.Get(ctx)
	router := handlers.NewRouter(log)

	config := config.Get()

	router.HandleFunc("/webhooks", Webhook)
	log.WithField("port", config.Port).Info("Starting web server")

	headersOk := muxhandlers.AllowedHeaders([]string{"X-Requested-With", "Origin", "Content-Type", "Accept", "Authorization"})
	originsOk := muxhandlers.AllowedOrigins([]string{"*"})
	methodsOk := muxhandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	err := http.ListenAndServe(fmt.Sprintf(":%v", config.Port), muxhandlers.CORS(originsOk, headersOk, methodsOk)(router))
	if err != nil {
		panic(err)
	}
}
