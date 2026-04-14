package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

func AddImageProcessing(ctx context.Context, client *firestore.Client) {
	log.Println("🔄 Images koleksiyonuna processing map'i ekleniyor...")

	imagesRef := client.Collection("images")
	iter := imagesRef.Documents(ctx)
	defer iter.Stop()

	bulkWriter := client.BulkWriter(ctx)
	count := 0
	skipped := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Doküman okuma hatası: %v", err)
		}

		data := doc.Data()

		// Eğer zaten processing varsa skip et
		if _, exists := data["processing"]; exists {
			skipped++
			continue
		}

		// Eski alanlardan değerleri al
		lastProcessedAt := data["last_processed_at"]
		retryCount := int64(0)
		if rc, ok := data["retry_count"].(int64); ok {
			retryCount = rc
		}

		// Processing map'i oluştur
		processingMap := map[string]interface{}{
			"last_processed_at": lastProcessedAt,
			"active_event_id":   uuid.New().String(),
			"retry_count":       retryCount,
			"status":            "processed",
			"version":           "v1",
		}

		// BulkWriter ile güncelle
		_, err = bulkWriter.Update(doc.Ref, []firestore.Update{
			{Path: "processing", Value: processingMap},
		})
		if err != nil {
			log.Fatalf("BulkWriter güncelleme hatası: %v", err)
		}

		count++
		if count%100 == 0 {
			log.Printf("  📝 %d image güncellendi", count)
		}
	}

	bulkWriter.End()

	log.Printf("✅ Toplam %d image'a processing map'i eklendi!", count)
	if skipped > 0 {
		log.Printf("⏭️  %d image zaten processing map'ine sahipti (skip edildi)", skipped)
	}
}
