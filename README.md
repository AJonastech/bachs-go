# bachs-go

A hand-written, idiomatic Go client for the [Bachs API](https://docs.bachs.io) — payments and billing infrastructure for internet businesses.

- **Typed everything** — resources, params, and errors are concrete Go types, not `map[string]any`.
- **Automatic retries** — 429s and 5xxs are retried with backoff, honoring `Retry-After`; writes retry only when made idempotent.
- **Auto-pagination** — range over every page with Go 1.23+ iterators; no manual cursor/offset bookkeeping.
- **No runtime dependencies** — standard library only (`net/http`, `encoding/json`).

> **Status:** early. The core client and the **Customers** and **Products** resources are complete and tested. The remaining resources follow the same pattern — see [Adding a resource](#adding-a-resource).

## Requirements

Go 1.25 or newer.

## Install

```sh
go get github.com/AJonastech/bachs-go
```

```go
import "github.com/AJonastech/bachs-go"
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AJonastech/bachs-go"
)

func main() {
	client, err := bachs.NewClient("sk_sandbox_...")
	if err != nil {
		log.Fatal(err)
	}

	cust, err := client.Customers.Create(context.Background(), &bachs.CustomerCreateParams{
		Email: "jane@example.com",
		Name:  bachs.String("Jane Doe"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cust.ID) // cust_...
}
```

Every method takes a `context.Context` first and a variadic tail of `RequestOption` last, so a call is always `(ctx, [id,] [params,] opts...)`.

A runnable end-to-end tour lives in [`examples/main.go`](examples/main.go):

```sh
BACHS_API_KEY=sk_sandbox_... go run ./examples
```

## Environments

The base URL is chosen automatically from your key prefix — **going live is a key swap, not a code change**:

| Key prefix    | Base URL                       |
| ------------- | ------------------------------ |
| `sk_sandbox_` | `https://sandbox-api.bachs.io` |
| `sk_live_`    | `https://api.bachs.io`         |

Point the client somewhere else (a proxy, or a mock in tests) with `WithBaseURL`:

```go
client, err := bachs.NewClient(key, bachs.WithBaseURL("https://proxy.internal"))
```

## Errors

Any non-2xx response comes back as an `*APIError`. Branch on the stable `ErrorCode` with the helpers rather than matching on message strings:

```go
_, err := client.Customers.Get(ctx, "cust_missing")
switch {
case bachs.IsNotFound(err):
	// no such customer
case bachs.IsValidation(err):
	// inspect apiErr.Fields for the offending fields
case bachs.IsRateLimited(err):
	// already retried with backoff; still rate limited
case err != nil:
	var apiErr *bachs.APIError
	if errors.As(err, &apiErr) {
		log.Printf("bachs %s (%d): %s [request %s]",
			apiErr.ErrorCode, apiErr.StatusCode, apiErr.Detail, apiErr.RequestID)
	}
}
```

`APIError` carries the HTTP `StatusCode`, the stable `ErrorCode`, a human-readable `Detail`, any per-field `Fields`, a `DocURL`, the `RequestID` (for support), and the raw body. Helpers: `IsValidation`, `IsNotFound`, `IsAuth`, `IsConflict`, `IsRateLimited`.

## Idempotency and retries

The client automatically retries idempotent requests (`GET`, `HEAD`, `DELETE`) on 429 and 5xx responses, honoring `Retry-After` with exponential backoff and full jitter, up to `WithMaxRetries` attempts (default 2).

Writes (`POST`, `PATCH`, `PUT`) are **only** retried when you attach an idempotency key — which is what makes the retry safe from duplicates:

```go
cust, err := client.Customers.Create(ctx, params,
	bachs.WithIdempotencyKey("order-42"))
```

Send the same key again and you get the original result back instead of a duplicate.

## Pagination

Each `List` returns a single `Page[T]`. To walk every page, range over the matching `All` method — it fetches pages as needed and yields one item at a time (Customers page by offset, Products by cursor; the iterator hides the difference):

```go
for prod, err := range client.Products.All(ctx, nil) {
	if err != nil {
		return err
	}
	fmt.Println(prod.ID)
}
```

Break out of the loop to stop early and no further requests are made. Use `bachs.Collect` to gather an iterator into a slice:

```go
products, err := bachs.Collect(client.Products.All(ctx, nil))
```

## Optional fields

- **Response** fields that may be absent are pointers (`*string`, `*time.Time`).
- **Request** fields that are optional are pointers too; set them inline with the helpers: `bachs.String`, `bachs.Int`, `bachs.Bool`, `bachs.MoneyPtr`, or the generic `bachs.Ptr`.
- **Money** is a decimal string (`"29.00"`), never minor units — see the `Money` type.
- **Tri-state** fields that must distinguish *leave unchanged* from *set to null* from *set to a value* use `Opt[T]`. The customer billing address on update is the canonical case:

```go
// leave the stored address untouched — omit the field entirely
client.Customers.Update(ctx, id, &bachs.CustomerUpdateParams{
	Name: bachs.String("New Name"),
})

// clear it
client.Customers.Update(ctx, id, &bachs.CustomerUpdateParams{
	BillingAddress: bachs.Null[bachs.Address](),
})

// replace it in full (any omitted component becomes null)
client.Customers.Update(ctx, id, &bachs.CustomerUpdateParams{
	BillingAddress: bachs.Set(bachs.Address{Line1: "40 Yaba Road", Country: "NG"}),
})
```

## Configuration

`NewClient` accepts functional options:

| Option                   | Purpose                                                       |
| ------------------------ | ------------------------------------------------------------ |
| `WithBaseURL(url)`       | Override the environment base URL (proxy, mock).             |
| `WithHTTPClient(c)`      | Supply your own `*http.Client` (custom timeout, transport).  |
| `WithHTTPHeader(k, v)`   | Add a default header sent on every request.                  |
| `WithMaxRetries(n)`      | Cap automatic retries (default 2).                           |
| `WithUserAgent(s)`       | Override the `User-Agent`.                                   |

Per-request options are the variadic tail on every method:

| Option                     | Purpose                                          |
| -------------------------- | ------------------------------------------------ |
| `WithIdempotencyKey(key)`  | Make a write safe to retry.                      |
| `WithRequestHeader(k, v)`  | Add a header to a single request.                |

A `Client` is safe for concurrent use by multiple goroutines.

## Resources

| Resource  | Methods                                                        |
| --------- | ------------------------------------------------------------- |
| Customers | `Create`, `Get`, `Update`, `List`, `All`                      |
| Products  | `Create`, `Get`, `Update`, `List`, `All`, `Archive`, `Unarchive` |

## Adding a resource

Every resource follows the same shape, so adding one (say, Refunds) is mechanical — copy [`customers.go`](customers.go) and adjust:

1. **Types** — a response struct (nullable fields as pointers), a `XCreateParams` / `XUpdateParams` (required fields as values, optional as pointers with `omitempty`), and an `XListParams` with a `query()` method.
2. **Service** — an `XService struct { core *client }` whose methods call `s.core.do(ctx, method, path, query, body, &out, opts)`.
3. **Auto-pager** — an `All` method that calls `paginate(...)`, advancing by offset or cursor as that resource requires.
4. **Wire it** — add the service field to `Client` and set it in `NewClient` (see [`bachs.go`](bachs.go)).
5. **Test it** — round-trip the request/response against `httptest`, following [`customers_test.go`](customers_test.go).

The shared core (auth, retries, idempotency, error decoding, pagination) is already handled — a new resource only describes its types and endpoints.

## Testing

The full suite runs offline against `httptest` — no API key or network needed:

```sh
go test ./...
```

## License

[MIT](LICENSE)
