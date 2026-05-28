package main

import (
	"time"

	"yoo/internal/models"
)

func newDemoSchema(fields ...models.RecordField) *models.RecordSchema {
	return &models.RecordSchema{Fields: fields}
}

func newDemoRecord(index int, data map[string]interface{}, status string, age time.Duration) *models.TemplateRecord {
	now := time.Now()
	return &models.TemplateRecord{
		ID:             int64(index),
		NoteTemplateID: 1,
		RecordIndex:    index,
		Data:           data,
		Status:         status,
		CreatedAt:      now.Add(-age),
		UpdatedAt:      now.Add(-age / 2),
	}
}
