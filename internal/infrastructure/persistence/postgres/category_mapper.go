package postgres

import (
	"app/internal/domain/category"
	"app/internal/infrastructure/persistence/models"
)

func toCategoryDomain(m *models.CategoryModel) *category.Category {
	c := &category.Category{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Name:      m.Name,
		Slug:      m.Slug,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.Description != nil {
		c.Description = *m.Description
	}
	if m.ImageURL != nil {
		c.ImageURL = *m.ImageURL
	}
	return c
}

func toCategoriesDomain(items []models.CategoryModel) []category.Category {
	result := make([]category.Category, len(items))
	for i, m := range items {
		result[i] = *toCategoryDomain(&m)
	}
	return result
}

func toCategoryModel(c *category.Category) *models.CategoryModel {
	m := &models.CategoryModel{
		ID:        c.ID,
		ParentID:  c.ParentID,
		Name:      c.Name,
		Slug:      c.Slug,
		SortOrder: c.SortOrder,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Description != "" {
		m.Description = &c.Description
	}
	if c.ImageURL != "" {
		m.ImageURL = &c.ImageURL
	}
	return m
}
