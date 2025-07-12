package model

// RevenueCatWebhook represents the top-level structure of a RevenueCat webhook payload.
type RevenueCatWebhook struct {
	APIVersion string          `json:"api_version"`
	Event      RevenueCatEvent `json:"event"`
}

// RevenueCatEvent represents the "event" object within a RevenueCat webhook.
type RevenueCatEvent struct {
	ID                       string   `json:"id"`
	Type                     string   `json:"type"`
	EventTimestampMS         int64    `json:"event_timestamp_ms"`
	AppUserID                string   `json:"app_user_id"`
	OriginalAppUserID        string   `json:"original_app_user_id"`
	Aliases                  []string `json:"aliases"`
	ProductID                string   `json:"product_id"`
	EntitlementIDs           []string `json:"entitlement_ids"`
	PeriodType               string   `json:"period_type"`
	PurchasedAtMS            int64    `json:"purchased_at_ms"`
	ExpirationAtMS           int64    `json:"expiration_at_ms"`
	TransactionID            string   `json:"transaction_id"`
	OriginalTransactionID    string   `json:"original_transaction_id"`
	IsFamilyShare            bool     `json:"is_family_share"`
	CountryCode              string   `json:"country_code"`
	Currency                 string   `json:"currency"`
	Price                    float64  `json:"price"`
	PriceInPurchasedCurrency float64  `json:"price_in_purchased_currency"`
	TakehomePercentage       float64  `json:"takehome_percentage"`
	Environment              string   `json:"environment"`
	ExpirationReason         *string  `json:"expiration_reason,omitempty"`
	TransferredFrom          []string `json:"transferred_from"`
	TransferredTo            []string `json:"transferred_to"`
	Store                    string   `json:"store"`
}
