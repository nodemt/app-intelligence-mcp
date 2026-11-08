package main

import (
	"app-intelligence-mcp/internal/mixpanel"
	"context"
	"log"
	"time"
)

func main() {
	oauthHandler, closeOAuth, err := mixpanel.NewLoopbackOAuth()
	if err != nil {
		log.Fatal(err)
	}
	defer closeOAuth()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	session, err := mixpanel.ConnectMCP(ctx, "https://mcp-eu.mixpanel.com/mcp", oauthHandler)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	log.Print("Connected to Mixpanel MCP")

	report, err := mixpanel.GetMCPReport(ctx, session, 4081780, 93902963)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("id=%d name=%q results=%s", report.ID, report.Name, report.Results)
}
