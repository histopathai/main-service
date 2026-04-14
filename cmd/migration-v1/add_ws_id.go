package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func AddWorkspaceIds(ctx context.Context, client *firestore.Client) {
	log.Println("🔄 Workspace ID'leri ekleniyor...")

	// 1. Patient ID -> Workspace ID mapping oluştur
	patientToWorkspace := make(map[string]string)

	patientsIter := client.Collection("patients").Documents(ctx)
	for {
		doc, err := patientsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Patient okuma hatası: %v", err)
		}

		data := doc.Data()
		patientID := doc.Ref.ID
		parentID, ok := data["parent_id"].(string)
		if ok && parentID != "" {
			patientToWorkspace[patientID] = parentID // parent_id = workspace_id
		}
	}
	patientsIter.Stop()

	log.Printf("📊 %d patient-workspace eşleşmesi bulundu", len(patientToWorkspace))

	// 2. Image'ları update et
	imageUpdateMap := make(map[string]string) // image_id -> workspace_id

	for patientID, workspaceID := range patientToWorkspace {
		// Bu patient'a ait image'ları bul
		imagesQuery := client.Collection("images").Where("parent_id", "==", patientID)
		imagesIter := imagesQuery.Documents(ctx)

		for {
			doc, err := imagesIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Image okuma hatası: %v", err)
			}

			imageID := doc.Ref.ID
			imageUpdateMap[imageID] = workspaceID
		}
		imagesIter.Stop()
	}

	log.Printf("📊 %d image güncellenecek", len(imageUpdateMap))

	// 3. BulkWriter ile image'ları güncelle
	bulkWriter := client.BulkWriter(ctx)
	count := 0

	for imageID, workspaceID := range imageUpdateMap {
		imageRef := client.Collection("images").Doc(imageID)
		_, err := bulkWriter.Update(imageRef, []firestore.Update{
			{Path: "ws_id", Value: workspaceID},
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
	log.Printf("✅ Toplam %d image'a ws_id eklendi!\n", count)

	// 4. Annotation'ları update et (varsa)
	updateAnnotations(ctx, client, imageUpdateMap)
}

func updateAnnotations(ctx context.Context, client *firestore.Client, imageToWorkspace map[string]string) {
	log.Println("🔄 Annotation'lara ws_id ekleniyor...")

	annotationUpdateMap := make(map[string]string) // annotation_id -> workspace_id

	for imageID, workspaceID := range imageToWorkspace {
		// Bu image'a ait annotation'ları bul
		annotationsQuery := client.Collection("annotations").Where("parent_id", "==", imageID)
		annotationsIter := annotationsQuery.Documents(ctx)

		for {
			doc, err := annotationsIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Annotation okuma hatası: %v", err)
			}

			annotationID := doc.Ref.ID
			annotationUpdateMap[annotationID] = workspaceID
		}
		annotationsIter.Stop()
	}

	log.Printf("📊 %d annotation güncellenecek", len(annotationUpdateMap))

	// BulkWriter ile annotation'ları güncelle
	bulkWriter := client.BulkWriter(ctx)
	count := 0

	for annotationID, workspaceID := range annotationUpdateMap {
		annotationRef := client.Collection("annotations").Doc(annotationID)
		_, err := bulkWriter.Update(annotationRef, []firestore.Update{
			{Path: "ws_id", Value: workspaceID},
		})
		if err != nil {
			log.Fatalf("BulkWriter güncelleme hatası: %v", err)
		}

		count++
		if count%100 == 0 {
			log.Printf("  📝 %d annotation güncellendi", count)
		}
	}

	bulkWriter.End()
	log.Printf("✅ Toplam %d annotation'a ws_id eklendi!\n", count)
}
