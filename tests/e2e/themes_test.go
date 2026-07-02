//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

const (
	themeModernBlueID   = "d1000000-0000-0000-0000-000000000001"
	themeMinimalLightID = "d1000000-0000-0000-0000-000000000002"
	themeBoldDarkID     = "d1000000-0000-0000-0000-000000000003"
)

func TestThemes_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/themes", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestThemes_ListAndPurchase(t *testing.T) {
	c := adminClient(t)

	listResp := c.do(http.MethodGet, "/api/v1/admin/themes", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID            string  `json:"id"`
			Slug          string  `json:"slug"`
			Price         float64 `json:"price"`
			IsPurchased   bool    `json:"is_purchased"`
			IsActiveTheme bool    `json:"is_active_theme"`
		} `json:"data"`
	}
	listResp.JSON(t, &list)
	if len(list.Data) < 3 {
		t.Fatalf("expected seeded themes, got %d", len(list.Data))
	}

	foundActive := false
	for _, th := range list.Data {
		if th.ID == themeModernBlueID && th.IsActiveTheme {
			foundActive = true
		}
		if th.Price == 0 && !th.IsPurchased {
			t.Fatalf("free theme %s should show is_purchased=true", th.Slug)
		}
	}
	if !foundActive {
		t.Fatal("expected modern-blue to be active theme")
	}

	purchaseResp := c.do(http.MethodPost, "/api/v1/admin/themes/"+themeMinimalLightID+"/purchase", nil, nil)
	purchaseResp.AssertStatus(t, http.StatusCreated)

	var purchase struct {
		ThemeID     string `json:"theme_id"`
		PurchasedAt string `json:"purchased_at"`
	}
	purchaseResp.JSON(t, &purchase)
	if purchase.ThemeID != themeMinimalLightID || purchase.PurchasedAt == "" {
		t.Fatalf("unexpected purchase: %+v", purchase)
	}

	againResp := c.do(http.MethodPost, "/api/v1/admin/themes/"+themeMinimalLightID+"/purchase", nil, nil)
	againResp.AssertStatus(t, http.StatusCreated)
}

func TestThemes_StoreStyle_RoundTrip(t *testing.T) {
	c := adminClient(t)
	store := storeClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/store-style", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var before struct {
		ID            string `json:"id"`
		ActiveThemeID *string `json:"active_theme_id"`
		FontFamily    string `json:"font_family"`
		Colors        struct {
			Primary string `json:"primary"`
		} `json:"colors"`
	}
	getResp.JSON(t, &before)
	if before.ID == "" {
		t.Fatal("expected store style id")
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/store-style", map[string]any{
		"active_theme_id": themeBoldDarkID,
		"font_family":     "Roboto",
		"colors": map[string]string{
			"primary":    "#111827",
			"background": "#ffffff",
		},
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		ActiveThemeID *string `json:"active_theme_id"`
		FontFamily    string  `json:"font_family"`
		ActiveTheme   *struct {
			Slug string `json:"slug"`
		} `json:"active_theme"`
		Colors struct {
			Primary    string `json:"primary"`
			Background string `json:"background"`
		} `json:"colors"`
	}
	updateResp.JSON(t, &updated)

	if updated.ActiveThemeID == nil || *updated.ActiveThemeID != themeBoldDarkID {
		t.Fatalf("active_theme_id = %v, want %s", updated.ActiveThemeID, themeBoldDarkID)
	}
	if updated.FontFamily != "Roboto" {
		t.Fatalf("font_family = %q, want Roboto", updated.FontFamily)
	}
	if updated.ActiveTheme == nil || updated.ActiveTheme.Slug != "bold-dark" {
		t.Fatalf("active_theme = %+v", updated.ActiveTheme)
	}
	if updated.Colors.Primary != "#111827" {
		t.Fatalf("primary = %q", updated.Colors.Primary)
	}
	if updated.Colors.Background != "#ffffff" {
		t.Fatalf("background = %q", updated.Colors.Background)
	}

	publicResp := store.do(http.MethodGet, "/api/v1/store/theme", nil, nil)
	publicResp.AssertStatus(t, http.StatusOK)

	var public struct {
		FontFamily string `json:"font_family"`
		Colors     struct {
			Primary string `json:"primary"`
		} `json:"colors"`
	}
	publicResp.JSON(t, &public)
	if public.FontFamily != "Roboto" {
		t.Fatalf("public font_family = %q", public.FontFamily)
	}
	if public.Colors.Primary != "#111827" {
		t.Fatalf("public primary = %q", public.Colors.Primary)
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodPut, "/api/v1/admin/store-style", map[string]any{
			"active_theme_id": themeModernBlueID,
			"font_family":     "Inter",
		}, nil)
	})
}

func TestThemes_Purchase_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPost, "/api/v1/admin/themes/00000000-0000-0000-0000-000000000099/purchase", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestThemes_StoreStyle_ValidationError(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPut, "/api/v1/admin/store-style", map[string]any{
		"active_theme_id": "not-a-uuid",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestThemes_StoreStyle_ThemeNotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPut, "/api/v1/admin/store-style", map[string]any{
		"active_theme_id": "00000000-0000-0000-0000-000000000099",
	}, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestThemes_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/themes", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestThemes_List_IncludesExpectedSlugs(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/themes", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			Slug string `json:"slug"`
		} `json:"data"`
	}
	resp.JSON(t, &list)

	want := map[string]bool{
		"modern-blue":   false,
		"minimal-light": false,
		"bold-dark":     false,
	}
	for _, th := range list.Data {
		if _, ok := want[th.Slug]; ok {
			want[th.Slug] = true
		}
	}
	for slug, found := range want {
		if !found {
			t.Fatalf("missing seeded theme slug %q", slug)
		}
	}
}

func TestThemes_StoreStyle_UpdateColorsOnly(t *testing.T) {
	c := adminClient(t)
	customPrimary := fmt.Sprintf("#%06x", testRunID&0xffffff)

	resp := c.do(http.MethodPut, "/api/v1/admin/store-style", map[string]any{
		"colors": map[string]string{
			"primary": customPrimary,
		},
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var style struct {
		Colors struct {
			Primary string `json:"primary"`
		} `json:"colors"`
	}
	resp.JSON(t, &style)
	if style.Colors.Primary != customPrimary {
		t.Fatalf("primary = %q, want %q", style.Colors.Primary, customPrimary)
	}
}
