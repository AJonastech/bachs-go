package bachs

import (
	"encoding/json"
	"time"
)

// Money is a monetary amount as a decimal string in the currency's major unit,
// e.g. "29.00" for $29.00 or "75000.00" for ₦75,000.00. The Bachs API never
// uses minor units (cents) and never accepts a bare number — always pair a
// Money value with its currency.
type Money string

// Metadata is arbitrary key/value data you attach to a resource. Bachs stores
// and returns it unchanged.
type Metadata map[string]any

// Common currency codes. The API restricts a product's primary currency to USD
// or NGN, with additional currencies available through currency options. These
// constants are provided for convenience; any string is accepted on the wire.
const (
	CurrencyUSD = "USD"
	CurrencyNGN = "NGN"
	CurrencyGHS = "GHS"
	CurrencyKES = "KES"
	CurrencyUGX = "UGX"
	CurrencyTZS = "TZS"
	CurrencyRWF = "RWF"
	CurrencyXAF = "XAF"
	CurrencyXOF = "XOF"
	CurrencyZMW = "ZMW"
)

// PriceType describes how a product is priced.
const (
	// PriceTypeFixed is a set amount, given in Amount.
	PriceTypeFixed = "fixed"
	// PriceTypeFree is no charge.
	PriceTypeFree = "free"
	// PriceTypeCustom lets the customer pay what they want, bounded by
	// MinimumAmount and MaximumAmount with an optional PresetAmount suggestion.
	PriceTypeCustom = "custom"
)

// Interval is the unit of time for a billing cycle or trial period.
const (
	IntervalDay   = "day"
	IntervalWeek  = "week"
	IntervalMonth = "month"
	IntervalYear  = "year"
)

// Address is a customer's billing address.
//
// It is treated as one atomic value: when supplied on an update it replaces the
// stored address in full rather than merging, so any component you leave empty
// becomes null. Line1 and Country are required whenever an address is supplied,
// and Country must be a valid ISO-3166-1 alpha-2 code (e.g. "NG", "FR"). See
// CustomerUpdateParams for how to leave an address untouched, replace it, or
// clear it.
type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	// Country is a two-letter ISO-3166-1 alpha-2 country code, e.g. "NG".
	Country string `json:"country,omitempty"`
}

// Price is the pricing on a product, as returned by the API.
type Price struct {
	Currency  string `json:"currency"`
	PriceType string `json:"price_type"`
	Amount    Money  `json:"amount"`
	// PresetAmount is the suggested amount prefilled at checkout for a custom
	// price. Non-nil only when PriceType is "custom".
	PresetAmount *Money `json:"preset_amount"`
	// MinimumAmount / MaximumAmount bound a custom price. Non-nil only when
	// PriceType is "custom".
	MinimumAmount *Money `json:"minimum_amount"`
	MaximumAmount *Money `json:"maximum_amount"`
	// CurrencyOptions holds prices in additional currencies.
	CurrencyOptions []CurrencyOption `json:"currency_options,omitempty"`
}

// CurrencyOption is a product price in an additional currency, as returned by
// the API.
type CurrencyOption struct {
	Currency      string `json:"currency"`
	Amount        Money  `json:"amount"`
	MinimumAmount *Money `json:"minimum_amount"`
	MaximumAmount *Money `json:"maximum_amount"`
}

// PriceParams sets the pricing on a product when creating it.
type PriceParams struct {
	// Currency is the product's primary currency (USD or NGN). Required.
	Currency string `json:"currency"`
	// PriceType is one of PriceTypeFixed, PriceTypeFree, or PriceTypeCustom.
	// Defaults to fixed when omitted.
	PriceType string `json:"price_type,omitempty"`
	// Amount is required when PriceType is fixed; omit for free and custom.
	Amount *Money `json:"amount,omitempty"`
	// PresetAmount, MinimumAmount, and MaximumAmount apply to custom pricing.
	PresetAmount    *Money                 `json:"preset_amount,omitempty"`
	MinimumAmount   *Money                 `json:"minimum_amount,omitempty"`
	MaximumAmount   *Money                 `json:"maximum_amount,omitempty"`
	CurrencyOptions []CurrencyOptionParams `json:"currency_options,omitempty"`
}

// CurrencyOptionParams sets a product price in one additional currency. Its
// currency cannot repeat the primary currency.
type CurrencyOptionParams struct {
	Currency      string `json:"currency"`
	Amount        *Money `json:"amount,omitempty"`
	PresetAmount  *Money `json:"preset_amount,omitempty"`
	MinimumAmount *Money `json:"minimum_amount,omitempty"`
	MaximumAmount *Money `json:"maximum_amount,omitempty"`
}

// UpdatePriceParams sets the price fields to change on a product update.
type UpdatePriceParams struct {
	Currency        string                 `json:"currency,omitempty"`
	PriceType       string                 `json:"price_type,omitempty"`
	Amount          *Money                 `json:"amount,omitempty"`
	PresetAmount    *Money                 `json:"preset_amount,omitempty"`
	MinimumAmount   *Money                 `json:"minimum_amount,omitempty"`
	MaximumAmount   *Money                 `json:"maximum_amount,omitempty"`
	CurrencyOptions []CurrencyOptionParams `json:"currency_options,omitempty"`
}

// Cadence describes how often something recurs: a Frequency of Interval units.
// For example {Interval: IntervalMonth, Frequency: 3} means every three months.
type Cadence struct {
	Interval  string `json:"interval,omitempty"`
	Frequency int    `json:"frequency,omitempty"`
}

// TrialPeriod is the length of a free trial before the first charge, e.g.
// {Interval: IntervalDay, Frequency: 14} for a 14-day trial.
type TrialPeriod struct {
	Interval  string `json:"interval"`
	Frequency int    `json:"frequency"`
}

// MediaItem is an image or file attached to a resource such as a product.
type MediaItem struct {
	ID            string    `json:"id"`
	URL           *string   `json:"url"`
	FileName      string    `json:"file_name"`
	MimeType      string    `json:"mime_type"`
	FileSizeBytes int       `json:"file_size_bytes"`
	CreatedAt     time.Time `json:"created_at"`
}

// Opt is a tri-state optional value for request parameters. It distinguishes
// three intentions that a plain pointer cannot express at once:
//
//   - unset   — the field is omitted from the request body entirely (leave the
//     stored value untouched). This is the zero value of Opt.
//   - null    — the field is sent as JSON null (clear the stored value).
//   - a value — the field is sent with that value.
//
// Construct one with Set or Null:
//
//	params.BillingAddress = bachs.Set(bachs.Address{Line1: "40 Yaba Road", Country: "NG"})
//	params.BillingAddress = bachs.Null[bachs.Address]() // clear it
//	// leaving params.BillingAddress as the zero Opt omits it
//
// Fields of type Opt must use the ",omitzero" JSON tag so an unset Opt is
// dropped from the encoded body.
type Opt[T any] struct {
	value   T
	present bool
	null    bool
}

// Set returns an Opt that sends v.
func Set[T any](v T) Opt[T] { return Opt[T]{value: v, present: true} }

// Null returns an Opt that sends JSON null, clearing the stored value.
func Null[T any]() Opt[T] { return Opt[T]{present: true, null: true} }

// IsZero reports whether the Opt is unset. It is used by encoding/json's
// omitzero tag option to omit unset fields.
func (o Opt[T]) IsZero() bool { return !o.present }

// Get returns the value and whether a non-null value is present.
func (o Opt[T]) Get() (T, bool) {
	return o.value, o.present && !o.null
}

// IsNull reports whether the Opt is explicitly set to null.
func (o Opt[T]) IsNull() bool { return o.present && o.null }

// MarshalJSON implements json.Marshaler.
func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if o.null {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

// UnmarshalJSON implements json.Unmarshaler so Opt can also round-trip in
// responses.
func (o *Opt[T]) UnmarshalJSON(data []byte) error {
	o.present = true
	if string(data) == "null" {
		o.null = true
		var zero T
		o.value = zero
		return nil
	}
	o.null = false
	return json.Unmarshal(data, &o.value)
}

// Ptr returns a pointer to v. It is a convenience for setting optional pointer
// fields on request params inline: bachs.Ptr("some value").
func Ptr[T any](v T) *T { return &v }

// String returns a pointer to v.
func String(v string) *string { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Int64 returns a pointer to v.
func Int64(v int64) *int64 { return &v }

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Float64 returns a pointer to v.
func Float64(v float64) *float64 { return &v }

// MoneyPtr returns a pointer to a Money value, for optional amount fields:
// bachs.MoneyPtr("29.00").
func MoneyPtr(v Money) *Money { return &v }
