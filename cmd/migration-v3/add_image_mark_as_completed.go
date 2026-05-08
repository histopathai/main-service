package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func AddImageMarkAsCompleted(ctx context.Context, client *firestore.Client) {
	collectionName := "images"
	log.Printf("🔄 %s koleksiyonu güncelleniyor...", collectionName)

	iter := client.Collection(collectionName).Documents(ctx)
	defer iter.Stop()

	bulkWriter := client.BulkWriter(ctx)
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Doküman okuma hatası: %v", err)
		}

		if _, ok := doc.Data()["marked_as_completed"]; ok {
			continue
		}

		_, err = bulkWriter.Update(doc.Ref, []firestore.Update{
			{Path: "marked_as_completed", Value: false},
		})
		if err != nil {
			log.Fatalf("BulkWriter güncelleme hatası: %v", err)
		}

		count++
		if count%100 == 0 {
			log.Printf("  📝 %d doküman kuyruğa eklendi", count)
		}
	}

	bulkWriter.End()

	log.Printf("✅ %s: Toplam %d doküman güncellendi!\n", collectionName, count)
}
