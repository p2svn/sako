package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"https://github.com/olahol/melody"
	rss "https://github.com/ungerik/go-rss"
)

func info(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"static/templates/layout.html",
		"static/templates/sidebar.html",
		"static/templates/info.html",
	)
	if err != nil {
		log.Print(err)
	}

	if err := t.Execute(w, "info"); err != nil {
		log.Print(err)
	}

	mel.HandleConnect(handleConnectInfo)
}

func updateInfo(s *melody.Session) error {
	graph, err := cryptoCompareGraph()
	if err != nil {
		return err
	}

	price, err := cryptoComparePrice()
	if err != nil {
		return err
	}

	feed, err := rss.Read("http://monero-observer.com/feed.rss")
	if err != nil {
		return err
	}

	msg, err := json.Marshal(struct {
		Type  string
		Price Price
		Graph Graph
		Feed  []rss.Item
	}{
		"info", price, graph, feed.Item,
	})
	if err != nil {
		return err
	}

	return s.Write(msg)
}

func handleConnectInfo(s *melody.Session) {
	if err := updateSidebar(s); err != nil {
		log.Println(err)
	}
	if err := updateInfo(s); err != nil {
		log.Println(err)
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
				if err := updateInfo(s); err != nil {
					log.Println(err)
				}
			}
		}
	}()
}
