package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// CSV columns: ws_id,patient_id,image_id,mpp,magnification
const mppMagnificationCSVPath = "/home/yasin/Projects/histopathai/ml/eda/mpp_magnification.csv"

type mppMagnificationEntry struct {
	WsID          string
	MPP           float64
	Magnification string
}

func loadMppMagnificationCSV(path string) map[string]mppMagnificationEntry {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("CSV açılamadı: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("CSV okuma hatası: %v", err)
	}

	result := make(map[string]mppMagnificationEntry)
	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		if len(row) < 5 {
			log.Printf("⚠️  satır %d: eksik kolon, atlanıyor", i+1)
			continue
		}

		wsID := strings.TrimSpace(row[0])
		imageID := strings.TrimSpace(row[2])
		mppStr := strings.TrimSpace(row[3])
		magStr := strings.TrimSpace(row[4])

		// Fields are optional per-image: a row missing either value is skipped.
		if imageID == "" || mppStr == "" || magStr == "" {
			continue
		}

		mpp, err := strconv.ParseFloat(mppStr, 64)
		if err != nil {
			log.Printf("⚠️  satır %d: geçersiz mpp değeri %q, atlanıyor", i+1, mppStr)
			continue
		}

		result[imageID] = mppMagnificationEntry{
			WsID:          wsID,
			MPP:           mpp,
			Magnification: magStr,
		}
	}
	return result
}

// BackfillMppMagnification reads the mpp/magnification CSV and backfills those two
// optional fields onto matching documents in the "images" collection. Images not
// present in the CSV are left untouched; documents already carrying "mpp" are
// skipped so the migration is safe to re-run.
func BackfillMppMagnification(ctx context.Context, client *firestore.Client) {
	log.Println("🔄 mpp / magnification_label CSV'den ingest ediliyor...")

	csvEntries := loadMppMagnificationCSV(mppMagnificationCSVPath)
	log.Printf("📊 CSV'de %d image satırı bulundu", len(csvEntries))

	matched := make(map[string]bool, len(csvEntries))

	collectionName := "images"
	iter := client.Collection(collectionName).Documents(ctx)
	defer iter.Stop()

	bulkWriter := client.BulkWriter(ctx)

	updated, skippedNotInCSV, skippedAlreadySet, wsMismatch := 0, 0, 0, 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Doküman okuma hatası: %v", err)
		}

		entry, ok := csvEntries[doc.Ref.ID]
		if !ok {
			skippedNotInCSV++
			continue
		}
		matched[doc.Ref.ID] = true

		data := doc.Data()
		if _, exists := data["mpp"]; exists {
			skippedAlreadySet++
			continue
		}

		if wsID, _ := data["ws_id"].(string); wsID != "" && wsID != entry.WsID {
			log.Printf("⚠️  image %s: ws_id uyuşmuyor (firestore=%s csv=%s), yine de güncelleniyor", doc.Ref.ID, wsID, entry.WsID)
			wsMismatch++
		}

		_, err = bulkWriter.Update(doc.Ref, []firestore.Update{
			{Path: "mpp", Value: entry.MPP},
			{Path: "magnification_label", Value: entry.Magnification},
		})
		if err != nil {
			log.Fatalf("BulkWriter güncelleme hatası (image %s): %v", doc.Ref.ID, err)
		}

		updated++
		if updated%200 == 0 {
			log.Printf("  📝 %d image kuyruğa eklendi", updated)
		}
	}

	bulkWriter.End()

	unmatched := 0
	for imageID := range csvEntries {
		if !matched[imageID] {
			unmatched++
			log.Printf("⚠️  CSV image_id %s Firestore'da bulunamadı", imageID)
		}
	}

	log.Printf(
		"✅ Tamamlandı: csv_satır=%d güncellendi=%d csv_de_yok_atlandı=%d zaten_ayarlı_atlandı=%d ws_id_uyumsuz=%d csv_eslesmedi=%d",
		len(csvEntries), updated, skippedNotInCSV, skippedAlreadySet, wsMismatch, unmatched,
	)
}
