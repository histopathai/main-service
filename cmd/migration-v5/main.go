package main

import (
	"context"
	"flag"
	"log"

	"cloud.google.com/go/firestore"
)

func main() {
	apply := flag.Bool("apply", false, "write the changes; without it only lists what would change")
	flag.Parse()
	ctx := context.Background()

	client, err := firestore.NewClientWithDatabase(ctx, "histopathai-478716", "(default)")
	if err != nil {
		log.Fatalf("Firestore client oluşturulamadı: %v", err)
	}
	defer client.Close()

	CloseUnsuitableRechecks(ctx, client, *apply)

	log.Println("🎉 Migration V5 tamamlandı")
}
