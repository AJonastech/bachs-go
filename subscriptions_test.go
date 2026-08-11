package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// TestSubscriptionCancel_NilParamsSendsNoBody guards the typed-nil-pointer body
// fix: Cancel(ctx, id, nil) must send an empty body, not the literal "null".
func TestSubscriptionCancel_NilParamsSendsNoBody(t *testing.T) {
	var gotMethod string
	var gotBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"sub_1","status":"canceled"}`)
	})

	sub, err := c.Subscriptions.Cancel(context.Background(), "sub_1", nil)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	if len(gotBody) != 0 {
		t.Errorf("body = %q, want empty (typed-nil params must not marshal to null)", gotBody)
	}
	if sub.Status != SubscriptionStatusCanceled {
		t.Errorf("status = %q, want canceled", sub.Status)
	}
}

func TestSubscriptionCancel_WithParams(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeJSON(t, r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"sub_1","status":"active","cancel_at_period_end":true}`)
	})

	sub, err := c.Subscriptions.Cancel(context.Background(), "sub_1", &SubscriptionCancelParams{
		CancelAtPeriodEnd: true,
		Reason:            String("too expensive"),
	})
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if gotBody["cancel_at_period_end"] != true || gotBody["reason"] != "too expensive" {
		t.Errorf("body = %v", gotBody)
	}
	if !sub.CancelAtPeriodEnd {
		t.Error("CancelAtPeriodEnd = false, want true")
	}
}

func TestSubscriptionList_QueryParams(t *testing.T) {
	var gotQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"items":[],"pagination":{"has_more":false}}`)
	})

	_, err := c.Subscriptions.List(context.Background(), &SubscriptionListParams{
		Limit:      10,
		CustomerID: "cust_1",
		Status:     SubscriptionStatusActive,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotQuery != "customer_id=cust_1&limit=10&status=active" {
		t.Errorf("query = %q", gotQuery)
	}
}
