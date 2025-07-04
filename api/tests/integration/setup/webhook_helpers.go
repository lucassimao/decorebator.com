package setup

import (
	"encoding/json"
	"os"

	"github.com/stripe/stripe-go/v82"
)

// CreateSignedStripeEvent creates a properly signed Stripe webhook event for testing
// Uses the STRIPE_WEBHOOK_SECRET environment variable for signature generation
// Automatically sets the correct API version for compatibility
func CreateSignedStripeEvent(event stripe.Event) ([]byte, string, error) {
	// Get webhook secret from environment
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		panic("STRIPE_WEBHOOK_SECRET env is not set")
	}

	// Set the correct API version to match stripe-go expectations
	event.APIVersion = "2025-04-30.basil"

	// Marshal the event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, "", err
	}

	// Generate the signature using Stripe's official testing utilities
	signedPayload := stripe.GenerateTestSignedPayload(&stripe.UnsignedPayload{
		Payload: payload,
		Secret:  secret,
	})

	// Use the signed payload and header from Stripe's testing utilities
	return signedPayload.Payload, signedPayload.Header, nil
}
