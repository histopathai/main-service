package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// CloseUnsuitableRechecks: images marked "Değerlendirmeye uygun değil" in Veri
// Etiketleyici before main-service kept Ek Kontrol in step left their requests
// open (or done with another outcome). Closes them with outcome "unsuitable",
// as main-service now does when the mark is set, so Ek Kontrol hides them.
func CloseUnsuitableRechecks(ctx context.Context, client *firestore.Client, apply bool) {
	iter := client.Collection("recheck_requests").Documents(ctx)
	defer iter.Stop()

	checked, changed := 0, 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("recheck_requests okunamadı: %v", err)
		}
		checked++
		data := doc.Data()
		status, _ := data["status"].(string)
		outcome, _ := data["outcome"].(string)
		if status == "done" && outcome == "unsuitable" {
			continue
		}

		image, err := client.Collection("images").Doc(doc.Ref.ID).Get(ctx)
		if err != nil {
			log.Printf("⚠️  %s: görüntü okunamadı, atlanıyor: %v", doc.Ref.ID, err)
			continue
		}
		img := image.Data()
		if unsuitable, _ := img["unsuitable"].(bool); !unsuitable {
			continue
		}
		by, _ := img["unsuitable_by"].(string)
		note, _ := img["unsuitable_note"].(string)
		at, ok := img["unsuitable_at"].(time.Time)
		if !ok {
			at = time.Now()
		}
		name, _ := data["image_name"].(string)
		log.Printf("→ %s (%s): %s/%s → done/unsuitable", doc.Ref.ID, name, status, outcome)
		changed++
		if !apply {
			continue
		}
		_, err = doc.Ref.Update(ctx, []firestore.Update{
			{Path: "status", Value: "done"},
			{Path: "outcome", Value: "unsuitable"},
			{Path: "completion_note", Value: note},
			{Path: "completed_by", Value: by},
			{Path: "completed_at", Value: at},
			{Path: "auto_completed", Value: false},
			{Path: "updated_at", Value: time.Now()},
		})
		if err != nil {
			log.Fatalf("%s güncellenemedi: %v", doc.Ref.ID, err)
		}
	}
	mode := "dry-run (yazmak için -apply)"
	if apply {
		mode = "yazıldı"
	}
	log.Printf("✅ %d talep tarandı, %d talep kapatılacak/kapatıldı — %s", checked, changed, mode)
}
