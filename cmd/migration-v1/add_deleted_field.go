package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func AddDeletedField(ctx context.Context, client *firestore.Client) {

	collections := []string{
		"workspaces",
		"patients",
		"images",
		"annotations",
		"annotation_types",
	}

	for _, collectionName := range collections {
		log.Printf("🔄 %s koleksiyonu güncelleniyor...", collectionName)

		collectionRef := client.Collection(collectionName)
		iter := collectionRef.Documents(ctx)

		// BulkWriter oluştur
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

			// BulkWriter ile güncelleme ekle
			_, err = bulkWriter.Update(doc.Ref, []firestore.Update{
				{Path: "is_deleted", Value: false},
			})
			if err != nil {
				log.Fatalf("BulkWriter güncelleme hatası: %v", err)
			}

			count++
			if count%100 == 0 {
				log.Printf("  📝 %d doküman kuyruğa eklendi", count)
			}
		}

		iter.Stop()

		// Tüm işlemleri flush et (commit)
		bulkWriter.End()

		log.Printf("✅ %s: Toplam %d doküman güncellendi!\n", collectionName, count)
	}

	log.Println("🎉 Tüm koleksiyonlar başarıyla güncellendi!")
}
