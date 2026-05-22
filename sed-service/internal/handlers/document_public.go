package handlers

import "github.com/industrial-sed/sed-service/internal/models"

func documentsPublic(list []models.Document) []models.DocumentPublic {
	out := make([]models.DocumentPublic, len(list))
	for i := range list {
		out[i] = models.ToDocumentPublic(&list[i])
	}
	return out
}
