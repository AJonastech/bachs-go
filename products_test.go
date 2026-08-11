package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestProductCreate_RequestAndResponse(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{
			"id": "prod_1",
			"name": "Pro Plan",
			"status": "active",
			"price": {"price_type":"fixed","currency":"USD","amount":"29.00"},
			"created_at": "2026-07-13T14:00:00.000Z",
			"updated_at": "2026-07-13T14:00:00.000Z",
			"billing_cycle": {"interval":"month","frequency":1},
			"total_amount": "0.00"
		}`)
	})

	prod, err := c.Products.Create(context.Background(), &ProductCreateParams{
		Name: "Pro Plan",
		Price: PriceParams{
			Currency:  CurrencyUSD,
			PriceType: PriceTypeFixed,
			Amount:    MoneyPtr("29.00"),
		},
		BillingCycle: &Cadence{Interval: IntervalMonth, Frequency: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Spot-check that the nested price serialized correctly.
	price, _ := gotBody["price"].(map[string]any)
	if price["currency"] != "USD" || price["amount"] != "29.00" {
		t.Errorf("request price = %v", price)
	}

	if prod.ID != "prod_1" || prod.Status != ProductStatusActive {
		t.Errorf("product = %+v", prod)
	}
	if prod.Price.Amount != "29.00" {
		t.Errorf("Price.Amount = %q, want 29.00", prod.Price.Amount)
	}
	if prod.BillingCycle == nil || prod.BillingCycle.Interval != IntervalMonth {
		t.Errorf("BillingCycle = %v", prod.BillingCycle)
	}
}

func TestProductArchiveUnarchive_Paths(t *testing.T) {
	cases := []struct {
		name     string
		call     func(c *Client) (*Product, error)
		wantPath string
	}{
		{
			name:     "archive",
			call:     func(c *Client) (*Product, error) { return c.Products.Archive(context.Background(), "prod_1") },
			wantPath: "/v1/products/prod_1/archive",
		},
		{
			name:     "unarchive",
			call:     func(c *Client) (*Product, error) { return c.Products.Unarchive(context.Background(), "prod_1") },
			wantPath: "/v1/products/prod_1/unarchive",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath string
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath = r.Method, r.URL.Path
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, `{"id":"prod_1","status":"archived"}`)
			})

			if _, err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if gotMethod != "POST" || gotPath != tc.wantPath {
				t.Errorf("request = %s %s, want POST %s", gotMethod, gotPath, tc.wantPath)
			}
		})
	}
}

func TestProductList_IncludeArchivedQuery(t *testing.T) {
	var gotQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"items":[],"pagination":{"has_more":false}}`)
	})

	_, err := c.Products.List(context.Background(), &ProductListParams{
		Limit:           10,
		IncludeArchived: true,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotQuery != "include_archived=true&limit=10" {
		t.Errorf("query = %q", gotQuery)
	}
}

// TestProductAll_WalksCursorPages verifies the auto-pager follows next_cursor
// and stops when it is null / has_more is false.
func TestProductAll_WalksCursorPages(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		w.WriteHeader(http.StatusOK)
		switch cursor {
		case "":
			_, _ = io.WriteString(w, `{
				"items":[{"id":"prod_1"}],
				"pagination":{"has_more":true,"next_cursor":"cur_1"}
			}`)
		default: // cursor=cur_1
			_, _ = io.WriteString(w, `{
				"items":[{"id":"prod_2"}],
				"pagination":{"has_more":false,"next_cursor":null}
			}`)
		}
	})

	ids, err := Collect(c.Products.All(context.Background(), nil))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(ids) != 2 || ids[0].ID != "prod_1" || ids[1].ID != "prod_2" {
		t.Errorf("ids = %+v, want prod_1, prod_2", ids)
	}
}

// TestProductAll_PropagatesError verifies an error mid-iteration is surfaced to
// the caller and stops the loop.
func TestProductAll_PropagatesError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"detail":"no scope","error_code":"FORBIDDEN"}`)
	})

	_, err := Collect(c.Products.All(context.Background(), nil))
	if !IsAuth(err) {
		t.Errorf("err = %v, want an auth error", err)
	}
}
