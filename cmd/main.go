package main

import (
	"fmt"
	"net/http"
	"noteservice/database"
	"noteservice/database/postgres"
	"noteservice/database/postgres/repository/noterepository"
	"noteservice/database/postgres/repository/userrepository"
	"noteservice/service/noteservice"
	passwordmanager "noteservice/service/passwordmanager"
	"noteservice/service/tokenservice"
	"noteservice/service/userservice"
	"noteservice/service/yandexspellerservice"
	"noteservice/transport/rest"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetReportCaller(true)
	log.SetLevel(logrus.DebugLevel)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := database.DbConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Dbname:   os.Getenv("DB_TABLE"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		Sslmode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := postgres.NewPostgresConnect(config)
	if err != nil {
		log.Fatalf("Cant connect to db: %v", err)
	}

	notes := noterepository.NewNoteStore(db)
	users := userrepository.NewUserStore(db)

	noteservice := noteservice.NewNoteService(notes)
	usersservice := userservice.NewUserService(users)
	speller := yandexspellerservice.NewSpeller()
	phasher := passwordmanager.NewPasswordHasher()
	tokens := tokenservice.NewTokenManager([]byte(os.Getenv("ACCESS_SECRET")), []byte(os.Getenv("REFRESH_SECRET")))

	handler := rest.NewHandler(log, usersservice, noteservice, tokens, phasher, speller)

	serv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", os.Getenv("URL"), os.Getenv("PORT")),
		Handler: handler.InitRouter(),
	}

	log.Infoln(serv.ListenAndServe())
}