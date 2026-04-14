package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"google.golang.org/api/iterator"
)

func CreateContents(ctx context.Context, client *firestore.Client) {
	log.Println("🔄 Content'ler oluşturuluyor...")

	imagesRef := client.Collection("images")
	contentsRef := client.Collection("contents")

	iter := imagesRef.Documents(ctx)
	defer iter.Stop()

	count := 0
	skipped := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Image okuma hatası: %v", err)
		}

		data := doc.Data()

		// Eğer zaten content_id'ler varsa skip et
		if _, exists := data["origin_content_id"]; exists {
			skipped++
			continue
		}

		// Gerekli alanları kontrol et
		originPath, ok := data["origin_path"].(string)
		if !ok || originPath == "" {
			log.Printf("⚠️  Image %s için origin_path bulunamadı, skip ediliyor", doc.Ref.ID)
			skipped++
			continue
		}

		name, _ := data["name"].(string)
		if name == "" {
			name = filepath.Base(originPath)
		}

		contentParentID := doc.Ref.ID // parent_id artık image'in kendi ID'si
		creatorID, _ := data["creator_id"].(string)
		size, _ := data["size"].(int64)
		processedPath, _ := data["processed_path"].(string)
		provider := string(vobj.ContentProviderGCS)

		ext := filepath.Ext(originPath)
		contentType := string(vobj.GetContentTypeFromExtension(ext))

		now := time.Now()

		// 1. Origin Content (orijinal dosya)
		originContentID := uuid.New().String()
		originContent := map[string]interface{}{
			"id":             originContentID,
			"name":           name,
			"entity_type":    string(vobj.EntityTypeContent),
			"parent_id":      contentParentID,
			"parent_type":    string(vobj.ParentTypeImage),
			"is_deleted":     false,
			"created_at":     now,
			"updated_at":     now,
			"creator_id":     creatorID,
			"provider":       provider,
			"content_type":   contentType,
			"path":           originPath,
			"upload_pending": false,
			"size":           size,
		}
		_, err = contentsRef.Doc(originContentID).Set(ctx, originContent)
		if err != nil {
			log.Fatalf("Origin content oluşturma hatası: %v", err)
		}

		// 2. DZI Content
		dziContentID := uuid.New().String()
		dziContent := map[string]interface{}{
			"id":             dziContentID,
			"name":           "image.dzi",
			"entity_type":    string(vobj.EntityTypeContent),
			"parent_id":      contentParentID,
			"parent_type":    string(vobj.ParentTypeImage),
			"is_deleted":     false,
			"created_at":     now,
			"updated_at":     now,
			"creator_id":     creatorID,
			"provider":       provider,
			"path":           fmt.Sprintf("%s/image.dzi", processedPath),
			"content_type":   string(vobj.ContentTypeApplicationDZI),
			"size":           int64(0),
			"upload_pending": false,
		}
		_, err = contentsRef.Doc(dziContentID).Set(ctx, dziContent)
		if err != nil {
			log.Fatalf("DZI content oluşturma hatası: %v", err)
		}

		// 3. Thumbnail Content
		thumbnailContentID := uuid.New().String()
		thumbnailContent := map[string]interface{}{
			"id":             thumbnailContentID,
			"name":           "thumbnail.jpg",
			"entity_type":    string(vobj.EntityTypeContent),
			"parent_id":      contentParentID,
			"parent_type":    string(vobj.ParentTypeImage),
			"is_deleted":     false,
			"created_at":     now,
			"updated_at":     now,
			"creator_id":     creatorID,
			"provider":       provider,
			"path":           fmt.Sprintf("%s/thumbnail.jpg", processedPath),
			"content_type":   string(vobj.ContentTypeThumbnailJPEG),
			"size":           int64(0),
			"upload_pending": false,
		}
		_, err = contentsRef.Doc(thumbnailContentID).Set(ctx, thumbnailContent)
		if err != nil {
			log.Fatalf("Thumbnail content oluşturma hatası: %v", err)
		}

		// 4. Tiles Content - DİZİN OLARAK (image_tiles/)
		// Eski versiyonda tiles bir dizine çıkarılıyordu, zip değildi
		tilesContentID := uuid.New().String()
		tilesContent := map[string]interface{}{
			"id":             tilesContentID,
			"name":           "image_files",
			"entity_type":    string(vobj.EntityTypeContent),
			"parent_id":      contentParentID,
			"parent_type":    string(vobj.ParentTypeImage),
			"is_deleted":     false,
			"created_at":     now,
			"updated_at":     now,
			"creator_id":     creatorID,
			"provider":       provider,
			"path":           fmt.Sprintf("%s/image_files/", processedPath), // Dizin olarak
			"content_type":   string(vobj.ContentTypeApplicationOctetStream),
			"size":           int64(0),
			"upload_pending": false,
		}
		_, err = contentsRef.Doc(tilesContentID).Set(ctx, tilesContent)
		if err != nil {
			log.Fatalf("Tiles content oluşturma hatası: %v", err)
		}

		// 5. Image dokümanını güncelle - content_id'leri ekle
		imageUpdates := []firestore.Update{
			{Path: "origin_content_id", Value: originContentID},
			{Path: "dzi_content_id", Value: dziContentID},
			{Path: "thumbnail_content_id", Value: thumbnailContentID},
			{Path: "tiles_content_id", Value: tilesContentID}, // tiles_content_id (zip değil)
		}
		_, err = imagesRef.Doc(doc.Ref.ID).Update(ctx, imageUpdates)
		if err != nil {
			log.Fatalf("Image güncelleme hatası: %v", err)
		}

		count++
		if count%10 == 0 {
			log.Printf("  📝 %d image için content'ler oluşturuldu", count)
		}
	}

	log.Printf("✅ Toplam %d image için content'ler oluşturuldu!", count)
	if skipped > 0 {
		log.Printf("⏭️  %d image skip edildi (zaten content_id var veya eksik alan)", skipped)
	}
}
