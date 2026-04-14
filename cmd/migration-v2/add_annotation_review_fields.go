package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func AddAnnotationReviewFields(ctx context.Context, client *firestore.Client) {
	collectionName := "annotations"
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

		data := doc.Data()
		updates := []firestore.Update{}

		// Eğer resource yoksa manual olarak ayarla
		if _, ok := data["resource"]; !ok {
			updates = append(updates, firestore.Update{Path: "resource", Value: "manual"})
		}

		// Eğer review_ids yoksa boş dizi olarak ayarla
		if _, ok := data["review_ids"]; !ok {
			updates = append(updates, firestore.Update{Path: "review_ids", Value: []string{}})
		}

		// Yalnızca güncellenecek alan varsa BulkWriter'a ekle
		if len(updates) > 0 {
			_, err = bulkWriter.Update(doc.Ref, updates)
			if err != nil {
				log.Fatalf("BulkWriter güncelleme hatası: %v", err)
			}
			count++
			if count%100 == 0 {
				log.Printf("  📝 %d doküman kuyruğa eklendi", count)
			}
		}
	}

	iter.Stop()

	// Tüm işlemleri flush et (commit)
	bulkWriter.End()

	log.Printf("✅ %s: Toplam %d doküman güncellendi!\n", collectionName, count)
	log.Println("🎉 Migration V2 başarıyla tamamlandı!")
}
