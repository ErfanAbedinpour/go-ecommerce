package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appattr "app/internal/application/attributedef"
	appval "app/internal/application/attributevalue"
	domainattr "app/internal/domain/attributedef"
	domainval "app/internal/domain/attributevalue"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.CatalogAttributeResponse{}

// AttributeHandler handles global product attribute HTTP endpoints.
type AttributeHandler struct {
	attrService *appattr.Service
	valService  *appval.Service
	validator   *validator.Validator
	log         *slog.Logger
}

// NewAttributeHandler creates a new AttributeHandler.
func NewAttributeHandler(attrService *appattr.Service, valService *appval.Service, v *validator.Validator, log *slog.Logger) *AttributeHandler {
	return &AttributeHandler{attrService: attrService, valService: valService, validator: v, log: log}
}

// --- Attribute definitions ---

// CreateAttribute godoc
// @Summary      Create product attribute
// @Description  Create a global product attribute definition.
// @Tags         product-attributes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateProductAttributeRequest  true  "Attribute data"
// @Success      201   {object}  dtoresponse.CatalogAttributeResponse
// @Router       /api/v1/admin/product-attributes [post]
func (h *AttributeHandler) CreateAttribute(w http.ResponseWriter, r *http.Request) {
	var req request.CreateProductAttributeRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	attr, err := h.attrService.Create(r.Context(), appattr.CreateInput{
		Name:      req.Name,
		Slug:      req.Slug,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToCatalogAttributeResponse(attr))
}

// ListAttributes godoc
// @Summary      List product attributes
// @Description  Get a paginated list of global attribute definitions.
// @Tags         product-attributes
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.CatalogAttributeListResponse
// @Router       /api/v1/admin/product-attributes [get]
func (h *AttributeHandler) ListAttributes(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseAttributeListFilter(r)

	result, err := h.attrService.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeListResponse(result))
}

// GetAttribute godoc
// @Summary      Get product attribute
// @Tags         product-attributes
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Attribute ID"
// @Success      200  {object}  dtoresponse.CatalogAttributeResponse
// @Router       /api/v1/admin/product-attributes/{id} [get]
func (h *AttributeHandler) GetAttribute(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	attr, err := h.attrService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeResponse(attr))
}

// UpdateAttribute godoc
// @Summary      Update product attribute
// @Tags         product-attributes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Attribute ID"
// @Param        body  body  request.UpdateProductAttributeRequest  true  "Attribute data"
// @Success      200   {object}  dtoresponse.CatalogAttributeResponse
// @Router       /api/v1/admin/product-attributes/{id} [put]
func (h *AttributeHandler) UpdateAttribute(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateProductAttributeRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	attr, err := h.attrService.Update(r.Context(), id, appattr.UpdateInput{
		Name:      req.Name,
		Slug:      req.Slug,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeResponse(attr))
}

// DeleteAttribute godoc
// @Summary      Delete product attribute
// @Tags         product-attributes
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Attribute ID"
// @Success      204  "No Content"
// @Router       /api/v1/admin/product-attributes/{id} [delete]
func (h *AttributeHandler) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.attrService.Delete(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// --- Attribute values ---

// CreateValue godoc
// @Summary      Create attribute value
// @Description  Create a value for a global product attribute.
// @Tags         product-attribute-values
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.CreateProductAttributeValueRequest  true  "Value data"
// @Success      201   {object}  dtoresponse.CatalogAttributeValueResponse
// @Router       /api/v1/admin/product-attribute-values [post]
func (h *AttributeHandler) CreateValue(w http.ResponseWriter, r *http.Request) {
	var req request.CreateProductAttributeValueRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	attrID, err := uuid.Parse(req.AttributeID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	val, err := h.valService.Create(r.Context(), appval.CreateInput{
		AttributeID: attrID,
		Value:       req.Value,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToCatalogAttributeValueResponse(val))
}

// ListValues godoc
// @Summary      List attribute values
// @Description  Get a paginated list of attribute values, optionally filtered by attribute_id.
// @Tags         product-attribute-values
// @Produce      json
// @Security     BearerAuth
// @Param        attribute_id  query  string  false  "Filter by attribute definition ID"
// @Success      200  {object}  dtoresponse.CatalogAttributeValueListResponse
// @Router       /api/v1/admin/product-attribute-values [get]
func (h *AttributeHandler) ListValues(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseAttributeValueListFilter(r)

	result, err := h.valService.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeValueListResponse(result))
}

// GetValue godoc
// @Summary      Get attribute value
// @Tags         product-attribute-values
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Value ID"
// @Success      200  {object}  dtoresponse.CatalogAttributeValueResponse
// @Router       /api/v1/admin/product-attribute-values/{id} [get]
func (h *AttributeHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	val, err := h.valService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeValueResponse(val))
}

// UpdateValue godoc
// @Summary      Update attribute value
// @Tags         product-attribute-values
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Value ID"
// @Param        body  body  request.UpdateProductAttributeValueRequest  true  "Value data"
// @Success      200   {object}  dtoresponse.CatalogAttributeValueResponse
// @Router       /api/v1/admin/product-attribute-values/{id} [put]
func (h *AttributeHandler) UpdateValue(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateProductAttributeValueRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	val, err := h.valService.Update(r.Context(), id, appval.UpdateInput{
		Value:     req.Value,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCatalogAttributeValueResponse(val))
}

// DeleteValue godoc
// @Summary      Delete attribute value
// @Tags         product-attribute-values
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Value ID"
// @Success      204  "No Content"
// @Router       /api/v1/admin/product-attribute-values/{id} [delete]
func (h *AttributeHandler) DeleteValue(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.valService.Delete(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

func parseAttributeListFilter(r *http.Request) domainattr.ListFilter {
	filter := domainattr.ListFilter{}
	if active := r.URL.Query().Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}
	return filter
}

func parseAttributeValueListFilter(r *http.Request) domainval.ListFilter {
	filter := domainval.ListFilter{}
	if attrID := r.URL.Query().Get("attribute_id"); attrID != "" {
		if id, err := uuid.Parse(attrID); err == nil {
			filter.AttributeID = &id
		}
	}
	if active := r.URL.Query().Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}
	return filter
}
