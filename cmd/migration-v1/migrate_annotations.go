package main

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// migrateAnnotations migrates annotations with the following changes:
// 1. tag_type: converts to lowercase
// 2. annotation_type_id: adds new field by matching annotation name with annotation_type name
// 3. tag_value -> value: migrates old field to new field
// Note: ws_id and is_deleted are handled in other migration files
//
// Also migrates annotation_types collection:
// - type (UPPER_CASE) -> tag_type (lower_case)
func MigrateAnnotations(ctx context.Context, client *firestore.Client) error {
	log.Println("🔄 Annotations migration başlatılıyor...")

	// 1. Migrate annotation_types: type -> tag_type (lowercase)
	log.Println("📝 Annotation types migration başlatılıyor...")
	annotationTypeMap := make(map[string]string)
	annotationTypesIter := client.Collection("annotation_types").Documents(ctx)
	defer annotationTypesIter.Stop()

	bulkWriterTypes := client.BulkWriter(ctx)
	typeCount := 0
	typeSkipped := 0

	for {
		doc, err := annotationTypesIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("⚠️  Annotation type okuma hatası: %v (skip)", err)
			typeSkipped++
			continue
		}

		data := doc.Data()
		name, nameOk := data["name"].(string)
		id, idOk := data["id"].(string)

		if nameOk && idOk && name != "" && id != "" {
			annotationTypeMap[name] = id
		}

		// Migrate type -> tag_type (lowercase)
		if typeValue, ok := data["type"].(string); ok && typeValue != "" {
			lowerType := strings.ToLower(typeValue)
			_, err = bulkWriterTypes.Update(doc.Ref, []firestore.Update{
				{Path: "tag_type", Value: lowerType},
			})
			if err != nil {
				log.Printf("⚠️  Annotation type güncelleme hatası (id: %s): %v - skip", doc.Ref.ID, err)
				typeSkipped++
				continue
			}
			typeCount++
		} else {
			typeSkipped++
		}
	}

	bulkWriterTypes.End()

	log.Printf("✅ %d annotation type güncellendi (type -> tag_type)", typeCount)
	if typeSkipped > 0 {
		log.Printf("⏭️  %d annotation type skip edildi", typeSkipped)
	}
	log.Printf("📊 %d annotation type mapping bulundu", len(annotationTypeMap))

	// 2. Migrate annotations
	annotationsRef := client.Collection("annotations")
	iter := annotationsRef.Documents(ctx)
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
			log.Printf("⚠️  Annotation okuma hatası: %v (skip)", err)
			skipped++
			continue
		}

		data := doc.Data()
		updates := []firestore.Update{}

		// 2.1. tag_type -> lowercase
		if tagType, ok := data["tag_type"].(string); ok && tagType != "" {
			lowerTagType := strings.ToLower(tagType)
			updates = append(updates, firestore.Update{
				Path:  "tag_type",
				Value: lowerTagType,
			})
		}

		// 2.2. annotation_type_id: lookup by name
		if name, ok := data["name"].(string); ok && name != "" {
			if annotationTypeID, found := annotationTypeMap[name]; found {
				updates = append(updates, firestore.Update{
					Path:  "annotation_type_id",
					Value: annotationTypeID,
				})
			} else {
				log.Printf("⚠️  Annotation %s için annotation_type bulunamadı (name: %s) - skip", doc.Ref.ID, name)
				skipped++
				continue
			}
		} else {
			log.Printf("⚠️  Annotation %s için name field yok - skip", doc.Ref.ID)
			skipped++
			continue
		}

		// 2.3. tag_value -> value
		if tagValue, exists := data["tag_value"]; exists {
			updates = append(updates, firestore.Update{
				Path:  "value",
				Value: tagValue, // any type (number or text)
			})
		}

		// Skip if no updates
		if len(updates) == 0 {
			skipped++
			continue
		}

		// Apply updates with BulkWriter
		_, err = bulkWriter.Update(doc.Ref, updates)
		if err != nil {
			log.Printf("⚠️  BulkWriter güncelleme hatası (annotation %s): %v - skip", doc.Ref.ID, err)
			skipped++
			continue
		}

		count++
		if count%100 == 0 {
			log.Printf("  📝 %d annotation güncellendi", count)
		}
	}

	bulkWriter.End()

	log.Printf("✅ Toplam %d annotation başarıyla migrate edildi!", count)
	if skipped > 0 {
		log.Printf("⏭️  %d annotation skip edildi", skipped)
	}

	return nil
}
