package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	client, err := firestore.NewClientWithDatabase(ctx, "histopathai-478716", "(default)")
	if err != nil {
		log.Fatalf("Firestore client oluşturulamadı: %v", err)
	}
	defer client.Close()

	AddDeletedField(ctx, client)
	AddImageProcessing(ctx, client)
	AddWorkspaceIds(ctx, client)

	CreateContents(ctx, client)
	MigrateAnnotations(ctx, client)
}
