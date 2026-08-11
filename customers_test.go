package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCustomerCreate_RequestAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{
			"customer_id": "cust_1",
			"email": "jane@example.com",
			"name": "Jane Doe",
			"metadata": {"plan":"pro"},
			"created_at": "2026-01-24T12:00:00.000Z"
		}`)
	})

	cust, err := c.Customers.Create(context.Background(), &CustomerCreateParams{
		Email: "jane@example.com",
		Name:  String("Jane Doe"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if gotMethod != "POST" || gotPath != "/v1/customers" {
		t.Errorf("request = %s %s, want POST /v1/customers", gotMethod, gotPath)
	}
	if gotBody["email"] != "jane@example.com" || gotBody["name"] != "Jane Doe" {
		t.Errorf("request body = %v", gotBody)
	}
	if cust.ID != "cust_1" {
		t.Errorf("ID = %q, want cust_1", cust.ID)
	}
	if cust.Email == nil || *cust.Email != "jane@example.com" {
		t.Errorf("Email = %v", cust.Email)
	}
	if cust.CreatedAt == nil || cust.CreatedAt.Year() != 2026 {
		t.Errorf("CreatedAt = %v, want a 2026 timestamp", cust.CreatedAt)
	}
	if cust.Metadata["plan"] != "pro" {
		t.Errorf("Metadata = %v", cust.Metadata)
	}
}

// TestCustomerUpdate_BillingAddressTriState verifies the three intents that
// Opt[Address] encodes: unset (field absent), null (field is JSON null), and a
// value (field is an object).
func TestCustomerUpdate_BillingAddressTriState(t *testing.T) {
	cases := []struct {
		name        string
		params      CustomerUpdateParams
		wantPresent bool
		wantNull    bool
	}{
		{
			name:        "unset omits the field",
			params:      CustomerUpdateParams{Name: String("New Name")},
			wantPresent: false,
		},
		{
			name:        "null clears the address",
			params:      CustomerUpdateParams{BillingAddress: Null[Address]()},
			wantPresent: true,
			wantNull:    true,
		},
		{
			name: "value replaces the address",
			params: CustomerUpdateParams{
				BillingAddress: Set(Address{Line1: "40 Yaba Road", Country: "NG"}),
			},
			wantPresent: true,
			wantNull:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal directly into a raw map to inspect exactly what goes on the
			// wire, since Opt controls presence via the omitzero tag.
			raw, err := json.Marshal(tc.params)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var m map[string]json.RawMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			field, present := m["billing_address"]
			if present != tc.wantPresent {
				t.Fatalf("billing_address present = %v, want %v (body: %s)", present, tc.wantPresent, raw)
			}
			if present && (string(field) == "null") != tc.wantNull {
				t.Errorf("billing_address = %s, wantNull = %v", field, tc.wantNull)
			}
		})
	}
}

func TestCustomerList_QueryParams(t *testing.T) {
	var gotQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"items":[],"pagination":{"has_more":false}}`)
	})

	_, err := c.Customers.List(context.Background(), &CustomerListParams{
		Limit:  25,
		Search: "jane@example.com",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotQuery != "limit=25&search=jane%40example.com" {
		t.Errorf("query = %q", gotQuery)
	}
}

// TestCustomerAll_WalksOffsetPages verifies the auto-pager advances by offset
// across multiple pages and stops when has_more is false.
func TestCustomerAll_WalksOffsetPages(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		w.WriteHeader(http.StatusOK)
		switch offset {
		case "", "0":
			_, _ = io.WriteString(w, `{
				"items":[{"customer_id":"cust_1"},{"customer_id":"cust_2"}],
				"pagination":{"has_more":true,"offset":0,"returned":2}
			}`)
		default: // offset=2
			_, _ = io.WriteString(w, `{
				"items":[{"customer_id":"cust_3"}],
				"pagination":{"has_more":false,"offset":2,"returned":1}
			}`)
		}
	})

	var ids []string
	for cust, err := range c.Customers.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, cust.ID)
	}

	want := []string{"cust_1", "cust_2", "cust_3"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, ids[i], want[i])
		}
	}
}

// TestCustomerAll_EarlyBreak verifies that breaking out of the range loop stops
// making requests.
func TestCustomerAll_EarlyBreak(t *testing.T) {
	var pages int
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pages++
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
			"items":[{"customer_id":"cust_1"}],
			"pagination":{"has_more":true,"offset":0,"returned":1}
		}`)
	})

	for cust, err := range c.Customers.All(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		_ = cust
		break // stop after the first item
	}
	if pages != 1 {
		t.Errorf("pages fetched = %d, want 1 after early break", pages)
	}
}
