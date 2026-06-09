package postgres

	import (
	"encoding/json"

	"app/internal/domain/product"
	"app/internal/infrastructure/persistence/models"
)

func toProductDomain(m *models.ProductModel) *product.Product {
	p := &product.Product{
		ID:         m.ID,
		CategoryID: m.CategoryID,
		Name:       m.Name,
		Slug:       m.Slug,
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
	for _, sku := range m.SKUs {
		p.SKUs = append(p.SKUs, toSkuDomain(sku))
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

func toAttributeDomain(m models.ProductAttributeModel) product.ProductAttribute {
	attr := product.ProductAttribute{
		ID:        m.ID,
		ProductID: m.ProductID,
		Name:      m.Name,
	}
	for _, v := range m.Values {
		attr.Values = append(attr.Values, product.ProductAttributeValue{
			ID:          v.ID,
			AttributeID: v.AttributeID,
			Value:       v.Value,
		})
	}
	return attr
}

func toSkuDomain(m models.SkuModel) product.Sku {
	sku := product.Sku{
		ID:        m.ID,
		ProductID: m.ProductID,
		Code:      m.Code,
		CreatedAt: m.CreatedAt,
	}
	
	var attrs map[string]string
	if m.Attributes != "" {
		_ = json.Unmarshal([]byte(m.Attributes), &attrs)
	}
	if attrs == nil {
		attrs = make(map[string]string)
	}
	sku.Attributes = attrs
	
	return sku
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
	for _, sku := range p.SKUs {
		m.SKUs = append(m.SKUs, toSkuModel(sku))
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

func toAttributeModel(attr product.ProductAttribute) models.ProductAttributeModel {
	m := models.ProductAttributeModel{
		ID:        attr.ID,
		ProductID: attr.ProductID,
		Name:      attr.Name,
	}
	for _, v := range attr.Values {
		m.Values = append(m.Values, models.ProductAttributeValueModel{
			ID:          v.ID,
			AttributeID: v.AttributeID,
			Value:       v.Value,
		})
	}
	return m
}

func toSkuModel(sku product.Sku) models.SkuModel {
	attrsJSON, _ := json.Marshal(sku.Attributes)
	return models.SkuModel{
		ID:         sku.ID,
		ProductID:  sku.ProductID,
		Code:       sku.Code,
		Attributes: string(attrsJSON),
		CreatedAt:  sku.CreatedAt,
	}
}
