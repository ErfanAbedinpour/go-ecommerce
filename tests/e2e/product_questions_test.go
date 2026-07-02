//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func questionBody(name, question string) map[string]any {
	return map[string]any{
		"asker_name":  name,
		"asker_email": uniqueEmail("qa-asker"),
		"question":    question,
	}
}

func TestProductQuestions_Ask_ValidationErrors(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing asker_name",
			body: map[string]any{
				"question": "Question without name?",
			},
		},
		{
			name: "missing question",
			body: map[string]any{
				"asker_name": "Guest",
			},
		},
		{
			name: "invalid email",
			body: map[string]any{
				"asker_name":  "Guest",
				"asker_email": "not-an-email",
				"question":    "Valid question?",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/questions", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestProductQuestions_Lifecycle_AskAnswerAndVisibility(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	askResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/questions",
		questionBody("E2E Asker", "Is this product available in blue?"), nil)
	askResp.AssertStatus(t, http.StatusCreated)

	var asked struct {
		ID        string `json:"id"`
		ProductID string `json:"product_id"`
		AskerName string `json:"asker_name"`
		Question  string `json:"question"`
		Status    string `json:"status"`
		Answer    string `json:"answer"`
	}
	askResp.JSON(t, &asked)
	if asked.ID == "" || asked.ProductID != productID {
		t.Fatalf("unexpected question: %+v", asked)
	}
	if asked.Status != "open" {
		t.Fatalf("status = %q, want open", asked.Status)
	}
	if asked.Answer != "" {
		t.Fatalf("answer = %q, want empty before admin response", asked.Answer)
	}

	openListResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/questions", nil, nil)
	openListResp.AssertStatus(t, http.StatusOK)

	var openList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	openListResp.JSON(t, &openList)
	for _, row := range openList.Data {
		if row.ID == asked.ID {
			t.Fatal("open question should not appear in public list until answered")
		}
	}

	c := adminClient(t)
	adminListResp := c.do(http.MethodGet, "/api/v1/admin/questions?status=open&product_id="+productID, nil, nil)
	adminListResp.AssertStatus(t, http.StatusOK)

	var adminList struct {
		Data []struct {
			ID       string `json:"id"`
			Question string `json:"question"`
		} `json:"data"`
	}
	adminListResp.JSON(t, &adminList)

	found := false
	for _, row := range adminList.Data {
		if row.ID == asked.ID {
			found = true
			if row.Question != asked.Question {
				t.Fatalf("question = %q", row.Question)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected open question in admin list")
	}

	answerResp := c.do(http.MethodPost, "/api/v1/admin/questions/"+asked.ID+"/answer", map[string]string{
		"answer": "Yes, blue is available while stock lasts.",
	}, nil)
	answerResp.AssertStatus(t, http.StatusNoContent)

	answeredListResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/questions?page=1&per_page=10", nil, nil)
	answeredListResp.AssertStatus(t, http.StatusOK)

	var answeredList struct {
		Data []struct {
			ID       string `json:"id"`
			Question string `json:"question"`
			Answer   string `json:"answer"`
			Status   string `json:"status"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	answeredListResp.JSON(t, &answeredList)

	foundPublic := false
	for _, row := range answeredList.Data {
		if row.ID == asked.ID {
			foundPublic = true
			if row.Status != "answered" {
				t.Fatalf("status = %q, want answered", row.Status)
			}
			if row.Answer != "Yes, blue is available while stock lasts." {
				t.Fatalf("answer = %q", row.Answer)
			}
			break
		}
	}
	if !foundPublic {
		t.Fatal("answered question should appear in public list")
	}
	if answeredList.Meta.Total == 0 {
		t.Fatal("expected total > 0 in public question list")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/questions/"+asked.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/questions", nil, nil)
	notFoundResp.AssertStatus(t, http.StatusOK)

	var afterDelete struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	notFoundResp.JSON(t, &afterDelete)
	for _, row := range afterDelete.Data {
		if row.ID == asked.ID {
			t.Fatal("deleted question should not appear in public list")
		}
	}
}

func TestProductQuestions_ProductNotFound(t *testing.T) {
	store := storeClient(t)
	missingID := "00000000-0000-0000-0000-000000000099"

	askResp := store.do(http.MethodPost, "/api/v1/store/products/"+missingID+"/questions",
		questionBody("Guest", "Does this exist?"), nil)
	askResp.AssertStatus(t, http.StatusNotFound)
	askResp.AssertErrorCode(t, "NOT_FOUND")

	listResp := store.do(http.MethodGet, "/api/v1/store/products/"+missingID+"/questions", nil, nil)
	listResp.AssertStatus(t, http.StatusNotFound)
}

func TestProductQuestions_Admin_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/questions", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestProductQuestions_Admin_AnswerValidation(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPost, "/api/v1/admin/questions/00000000-0000-0000-0000-000000000099/answer", map[string]string{}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestProductQuestions_Admin_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/questions", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestProductQuestions_Admin_ListWithSearch(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	uniqueQ := fmt.Sprintf("E2E unique question marker %d?", testRunID)
	askResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/questions",
		questionBody("Search Tester", uniqueQ), nil)
	askResp.AssertStatus(t, http.StatusCreated)

	var question struct {
		ID string `json:"id"`
	}
	askResp.JSON(t, &question)

	c := adminClient(t)
	searchResp := c.do(http.MethodGet, "/api/v1/admin/questions?q=unique+question+marker&product_id="+productID, nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var results struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	searchResp.JSON(t, &results)

	found := false
	for _, row := range results.Data {
		if row.ID == question.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected question in admin search results")
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/questions/"+question.ID, nil, nil)
	})
}
