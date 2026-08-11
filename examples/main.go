// Command examples runs a short end-to-end tour of the bachs SDK against the
// sandbox. It creates a customer and a product, updates the customer, and lists
// products, printing what it does at each step.
//
// Run it with a sandbox secret key:
//
//	BACHS_API_KEY=sk_sandbox_... go run ./examples
//
// Nothing here talks to production: a sk_sandbox_ key selects the sandbox base
// URL automatically.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AJonastech/bachs-go"
)

func main() {
	apiKey := os.Getenv("BACHS_API_KEY")
	if apiKey == "" {
		log.Fatal("set BACHS_API_KEY to a sk_sandbox_... key to run this example")
	}

	client, err := bachs.NewClient(apiKey)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	// Give the whole tour a deadline; every call takes the same context.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := run(ctx, client); err != nil {
		// Pull structured detail out of any API error for a useful message.
		var apiErr *bachs.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("bachs %s (%d): %s [request %s]",
				apiErr.ErrorCode, apiErr.StatusCode, apiErr.Detail, apiErr.RequestID)
		}
		log.Fatal(err)
	}
}

func run(ctx context.Context, client *bachs.Client) error {
	// 1. Create a customer. The idempotency key makes this create safe to retry:
	// send the same key again and you get the original customer, not a duplicate.
	cust, err := client.Customers.Create(ctx,
		&bachs.CustomerCreateParams{
			Email:       "jane@example.com",
			Name:        bachs.String("Jane Doe"),
			PhoneNumber: bachs.String("+2348012345678"),
			Metadata:    bachs.Metadata{"signup_source": "example"},
		},
		bachs.WithIdempotencyKey("example-customer-jane"),
	)
	if err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	fmt.Printf("created customer %s (%s)\n", cust.ID, deref(cust.Email))

	// 2. Update the customer's billing address. BillingAddress is tri-state:
	// Set replaces it in full, Null would clear it, and leaving it unset would
	// leave the stored address untouched.
	cust, err = client.Customers.Update(ctx, cust.ID, &bachs.CustomerUpdateParams{
		BillingAddress: bachs.Set(bachs.Address{
			Line1:   "40 Yaba Road",
			City:    "Lagos",
			Country: "NG",
		}),
	})
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	fmt.Printf("updated billing address for %s\n", cust.ID)

	// 3. Create a one-time product priced at $29.00. Add a BillingCycle to make
	// it recurring instead.
	prod, err := client.Products.Create(ctx,
		&bachs.ProductCreateParams{
			Name:        "Pro Plan",
			Description: bachs.String("Everything in Basic, plus priority support"),
			Price: bachs.PriceParams{
				Currency:  bachs.CurrencyUSD,
				PriceType: bachs.PriceTypeFixed,
				Amount:    bachs.MoneyPtr("29.00"),
			},
		},
		bachs.WithIdempotencyKey("example-product-pro"),
	)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	fmt.Printf("created product %s at %s %s\n", prod.ID, prod.Price.Currency, prod.Price.Amount)

	// 4. List products with the auto-pager, which follows the cursor across
	// pages. Stop after a few so the example stays short.
	fmt.Println("products:")
	shown := 0
	for p, err := range client.Products.All(ctx, &bachs.ProductListParams{Limit: 20}) {
		if err != nil {
			return fmt.Errorf("list products: %w", err)
		}
		fmt.Printf("  - %s  %s  %s %s\n", p.ID, p.Name, p.Price.Currency, p.Price.Amount)
		if shown++; shown >= 5 {
			break
		}
	}

	return nil
}

// deref returns the pointed-to string, or "" if the pointer is nil.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
