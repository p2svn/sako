package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"https://github.com/olahol/melody"
	"https://github.com/onodera-punpun/sako/monero"
)

func history(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"static/templates/layout.html",
		"static/templates/sidebar.html",
		"static/templates/history.html",
	)
	if err != nil {
		log.Print(err)
	}

	if err := t.Execute(w, "history"); err != nil {
		log.Print(err)
	}

	mel.HandleConnect(handleConnectHistory)
}

func updateHistory(s *melody.Session) error {
	price, err := cryptoComparePrice()
	if err != nil {
		return err
	}

	transfers, err := wallet.Transfers(true, true, true, true)
	if err != nil {
		return err
	}

	msg, err := json.Marshal(struct {
		Type      string
		Price     Price
		Transfers []monero.Transfer
	}{
		"history", price, transfers,
	})
	if err != nil {
		return err
	}

	return s.Write(msg)
}

func handleConnectHistory(s *melody.Session) {
	if err := updateSidebar(s); err != nil {
		log.Println(err)
	}
	if err := updateHistory(s); err != nil {
		return
	}

	go func() {
		fastTicker := time.NewTicker(5 * time.Second)
		slowTicker := time.NewTicker(20 * time.Second)
		defer func() {
			fastTicker.Stop()
			slowTicker.Stop()
			s.Close()
		}()

		for {
			if s.IsClosed() {
				return
			}

			select {
			case <-fastTicker.C:
				if err := updateSidebar(s); err != nil {
					log.Println(err)
				}
			case <-slowTicker.C:
				if err := updateHistory(s); err != nil {
					log.Println(err)
				}
			}
		}
	}()
}
