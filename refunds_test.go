package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestRefundCreate_RequestAndResponse(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{
			"refund_id":"ref_1","charge_id":"chg_1","reference":"r-1",
			"status":"processing","requested_amount":"10.00","refund_fee_amount":"0.00",
			"fee_bearer":"org","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"
		}`)
	})

	ref, err := c.Refunds.Create(context.Background(), &RefundCreateParams{
		ChargeID:  "chg_1",
		Reference: "r-1",
		Amount:    MoneyPtr("10.00"),
		FeeBearer: String(FeeBearerOrg),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if gotBody["charge_id"] != "chg_1" || gotBody["amount"] != "10.00" || gotBody["fee_bearer"] != "org" {
		t.Errorf("request body = %v", gotBody)
	}
	if ref.RefundID != "ref_1" || ref.Status != RefundStatusProcessing {
		t.Errorf("refund = %+v", ref)
	}
}

// TestRefundList_SynthesizesPagination checks that the non-standard {total,
// items} envelope is adapted into the SDK's standard Page/ListMeta, with a
// HasMore derived from offset + returned < total.
func TestRefundList_SynthesizesPagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"total":3,"items":[
			{"refund_id":"ref_1"},{"refund_id":"ref_2"}
		]}`)
	})

	page, err := c.Refunds.List(context.Background(), &RefundListParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(page.Items))
	}
	if page.Pagination.Total != 3 || page.Pagination.Returned != 2 {
		t.Errorf("meta = %+v", page.Pagination)
	}
	if !page.Pagination.HasMore {
		t.Error("HasMore = false, want true (0 + 2 < 3)")
	}
}

// TestRefundAll_WalksOffsetPages verifies the derived HasMore drives the pager
// to a natural stop when the running offset reaches total.
func TestRefundAll_WalksOffsetPages(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		w.WriteHeader(http.StatusOK)
		switch offset {
		case "", "0":
			_, _ = io.WriteString(w, `{"total":3,"items":[{"refund_id":"ref_1"},{"refund_id":"ref_2"}]}`)
		default: // offset=2
			_, _ = io.WriteString(w, `{"total":3,"items":[{"refund_id":"ref_3"}]}`)
		}
	})

	got, err := Collect(c.Refunds.All(context.Background(), &RefundListParams{Limit: 2}))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(got) != 3 || got[0].RefundID != "ref_1" || got[2].RefundID != "ref_3" {
		t.Errorf("refunds = %+v, want ref_1..ref_3", got)
	}
}
