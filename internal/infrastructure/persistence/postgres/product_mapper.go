package postgres

import (
	"app/internal/domain/product"
	"app/internal/infrastructure/persistence/models"
)

func toProductDomain(m *models.ProductModel) *product.Product {
	p := &product.Product{
		ID:         m.ID,
		CategoryID: m.CategoryID,
		Name:       m.Name,
		Slug:       m.Slug,
		SKU:        m.SKU,
		Price:      m.Price,
		SalePrice:  m.SalePrice,
		IsFeatured: m.IsFeatured,
		Status:     product.Status(m.Status),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.Description != nil {
		p.Description = *m.Description
	}
	if m.ShortDescription != nil {
		p.ShortDescription = *m.ShortDescription
	}
	if m.Brand != nil {
		p.Brand = *m.Brand
	}

	for _, img := range m.Images {
		p.Images = append(p.Images, toImageDomain(img))
	}
	for _, attr := range m.Attributes {
		p.Attributes = append(p.Attributes, toAttributeDomain(attr))
	}
	if m.Inventory != nil {
		p.Inventory = toInventoryDomain(*m.Inventory)
	}

	return p
}

func toProductsDomain(items []models.ProductModel) []product.Product {
	result := make([]product.Product, len(items))
	for i, m := range items {
		result[i] = *toProductDomain(&m)
	}
	return result
}

func toImageDomain(m models.ProductImageModel) product.Image {
	img := product.Image{
		ID:        m.ID,
		ProductID: m.ProductID,
		URL:       m.URL,
		SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt,
	}
	if m.AltText != nil {
		img.AltText = *m.AltText
	}
	return img
}

func toAttributeDomain(m models.ProductAttributeModel) product.Attribute {
	return product.Attribute{
		ID:        m.ID,
		ProductID: m.ProductID,
		Name:      m.Name,
		Value:     m.Value,
	}
}

func toInventoryDomain(m models.InventoryModel) product.Inventory {
	return product.Inventory{
		ID:                m.ID,
		ProductID:         m.ProductID,
		Quantity:          m.Quantity,
		LowStockThreshold: m.LowStockThreshold,
		UpdatedAt:         m.UpdatedAt,
	}
}

func toProductModel(p *product.Product) *models.ProductModel {
	m := &models.ProductModel{
		ID:         p.ID,
		CategoryID: p.CategoryID,
		Name:       p.Name,
		Slug:       p.Slug,
		SKU:        p.SKU,
		Price:      p.Price,
		SalePrice:  p.SalePrice,
		IsFeatured: p.IsFeatured,
		Status:     p.Status.String(),
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	if p.Description != "" {
		m.Description = &p.Description
	}
	if p.ShortDescription != "" {
		m.ShortDescription = &p.ShortDescription
	}
	if p.Brand != "" {
		m.Brand = &p.Brand
	}

	for _, img := range p.Images {
		m.Images = append(m.Images, toImageModel(img))
	}
	for _, attr := range p.Attributes {
		m.Attributes = append(m.Attributes, toAttributeModel(attr))
	}
	m.Inventory = &models.InventoryModel{
		ID:                p.Inventory.ID,
		ProductID:         p.Inventory.ProductID,
		Quantity:          p.Inventory.Quantity,
		LowStockThreshold: p.Inventory.LowStockThreshold,
		UpdatedAt:         p.Inventory.UpdatedAt,
	}

	return m
}

func toImageModel(img product.Image) models.ProductImageModel {
	m := models.ProductImageModel{
		ID:        img.ID,
		ProductID: img.ProductID,
		URL:       img.URL,
		SortOrder: img.SortOrder,
		CreatedAt: img.CreatedAt,
	}
	if img.AltText != "" {
		m.AltText = &img.AltText
	}
	return m
}

func toAttributeModel(attr product.Attribute) models.ProductAttributeModel {
	return models.ProductAttributeModel{
		ID:        attr.ID,
		ProductID: attr.ProductID,
		Name:      attr.Name,
		Value:     attr.Value,
	}
}
