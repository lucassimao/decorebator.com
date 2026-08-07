package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/config"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/security"
	"decorebator.com/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"golang.org/x/oauth2/google"
)

const googleAndroidPublisherScope = "https://www.googleapis.com/auth/androidpublisher"

// Context holds all application services and dependencies
type Context struct {
	// Core dependencies
	Database    *pgxpool.Pool
	RedisClient *redis.Client
	JobService  service.JobService

	// Services
	WordService                  *service.WordService
	WordlistService              *service.WordlistService
	UserService                  *service.UserService
	AuthTokens                   *service.AccessTokenService
	AuthSessions                 *service.AuthSessionService
	DefinitionService            *service.DefinitionService
	DefinitionImageService       *service.DefinitionImageService
	SubscriptionService          *service.SubscriptionService
	RevenueCatService            service.RevenueCatService
	LeitnerTrackingService       *service.LeitnerTrackingService
	LeitnerSystemStrategy        *service.LeitnerSystemStrategy
	ErrorReportService           *service.ErrorReportService
	AnalyticsService             service.AnalyticsServiceInterface
	RealtimeTelemetryService     *service.RealtimeTelemetryService
	MailService                  *mail.MailService
	StoreEntitlementRepository   *repository.StoreEntitlementRepository
	ProviderEventInboxRepository *repository.ProviderEventInboxRepository
	AppleNotificationIngestor    *service.AppleNotificationIngestor
	GoogleRTDNIngestor           *service.GoogleRTDNIngestor
	EffectiveAccessService       *service.EffectiveAccessService
	StoreIAPOrchestrator         *service.StoreIAPOrchestrator
	GoogleAcknowledgementWorker  *service.GoogleAcknowledgementRetryWorker
	GoogleAcknowledgementSweep   *service.GoogleAcknowledgementSweepWorker
	StoreIAPRequestLimiter       *service.StoreIAPRequestLimiter
	StoreIAPEnabled              bool
	LegacyProviderSurfaceEnabled bool
	LegacyRevenueCatWebhookAuth  string

	// Monitoring
	// Configuration
	Environment string
}

// ContextBuilder provides a fluent API for building Context
type ContextBuilder struct {
	context                  *Context
	errors                   []error
	revenueCatServiceFactory func(db *pgxpool.Pool) service.RevenueCatService
	legacyProviderConfig     config.LegacyProviderConfig
}

// NewContext creates a new Context builder
func NewContext() *ContextBuilder {
	return &ContextBuilder{
		context: &Context{},
		errors:  make([]error, 0),
	}
}

// WithDatabase sets the database connection
func (b *ContextBuilder) WithDatabase(db *pgxpool.Pool) *ContextBuilder {
	if db == nil {
		b.errors = append(b.errors, errors.New("database connection cannot be nil"))
		return b
	}
	b.context.Database = db
	return b
}

// WithRedisClient sets the Redis client
func (b *ContextBuilder) WithRedisClient(client *redis.Client) *ContextBuilder {
	if client == nil {
		b.errors = append(b.errors, errors.New("Redis client cannot be nil"))
		return b
	}
	b.context.RedisClient = client
	return b
}

// WithJobService sets the JobService for background job operations
func (b *ContextBuilder) WithJobService(jobService service.JobService) *ContextBuilder {
	if jobService == nil {
		b.errors = append(b.errors, errors.New("job service cannot be nil"))
		return b
	}
	b.context.JobService = jobService
	return b
}

// WithEnvironment sets the application environment
func (b *ContextBuilder) WithEnvironment(env string) *ContextBuilder {
	b.context.Environment = env
	return b
}

// WithRevenueCatService sets a custom RevenueCat service
func (b *ContextBuilder) WithRevenueCatService(revenueCatService service.RevenueCatService) *ContextBuilder {
	b.context.RevenueCatService = revenueCatService
	return b
}

// WithWordService sets a custom word service
func (b *ContextBuilder) WithWordService(wordService *service.WordService) *ContextBuilder {
	b.context.WordService = wordService
	return b
}

// WithWordlistService sets a custom wordlist service
func (b *ContextBuilder) WithWordlistService(wordlistService *service.WordlistService) *ContextBuilder {
	b.context.WordlistService = wordlistService
	return b
}

// WithDefinitionService sets a custom definition service
func (b *ContextBuilder) WithDefinitionService(definitionService *service.DefinitionService) *ContextBuilder {
	b.context.DefinitionService = definitionService
	return b
}

// WithDefinitionImageService sets a custom definition image service
func (b *ContextBuilder) WithDefinitionImageService(definitionImageService *service.DefinitionImageService) *ContextBuilder {
	b.context.DefinitionImageService = definitionImageService
	return b
}

// WithUserService sets a custom user service
func (b *ContextBuilder) WithUserService(userService *service.UserService) *ContextBuilder {
	b.context.UserService = userService
	return b
}

func (b *ContextBuilder) WithAuthTokens(authTokens *service.AccessTokenService) *ContextBuilder {
	b.context.AuthTokens = authTokens
	return b
}

// WithSubscriptionService sets a custom subscription service
func (b *ContextBuilder) WithSubscriptionService(subscriptionService *service.SubscriptionService) *ContextBuilder {
	b.context.SubscriptionService = subscriptionService
	return b
}

// WithRevenueCatServiceFunc sets a custom RevenueCat service using a factory function
func (b *ContextBuilder) WithRevenueCatServiceFunc(factory func(db *pgxpool.Pool) service.RevenueCatService) *ContextBuilder {
	b.revenueCatServiceFactory = factory
	return b
}

// WithLeitnerSystemStrategy sets a custom Leitner system strategy
func (b *ContextBuilder) WithLeitnerSystemStrategy(strategy *service.LeitnerSystemStrategy) *ContextBuilder {
	b.context.LeitnerSystemStrategy = strategy
	return b
}

// WithErrorReportService sets a custom error report service
func (b *ContextBuilder) WithErrorReportService(errorReportService *service.ErrorReportService) *ContextBuilder {
	b.context.ErrorReportService = errorReportService
	return b
}

// WithAnalyticsService sets a custom analytics service
func (b *ContextBuilder) WithAnalyticsService(analyticsService service.AnalyticsServiceInterface) *ContextBuilder {
	b.context.AnalyticsService = analyticsService
	return b
}

// WithRealtimeTelemetryService sets a custom realtime telemetry service
func (b *ContextBuilder) WithRealtimeTelemetryService(telemetryService *service.RealtimeTelemetryService) *ContextBuilder {
	b.context.RealtimeTelemetryService = telemetryService
	return b
}

// Build constructs the Context with all dependencies initialized
func (b *ContextBuilder) Build() (*Context, error) {
	// Check for builder errors first
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("Context validation failed: %v", b.errors)
	}

	// Validate required dependencies
	if b.context.Database == nil {
		return nil, errors.New("database connection is required")
	}
	legacyProviderConfig, err := config.LoadLegacyProviderConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to configure legacy provider surface: %w", err)
	}
	b.legacyProviderConfig = legacyProviderConfig
	b.context.LegacyProviderSurfaceEnabled = legacyProviderConfig.Enabled
	b.context.LegacyRevenueCatWebhookAuth = legacyProviderConfig.RevenueCatWebhookAuthorization
	if b.context.AuthTokens == nil {
		authConfig, authErr := config.LoadAuthConfig()
		if authErr != nil {
			return nil, fmt.Errorf("failed to configure authentication: %w", authErr)
		}
		b.context.AuthTokens, err = service.NewAccessTokenService(authConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize authentication: %w", err)
		}
	}
	if b.context.AuthSessions == nil {
		b.context.AuthSessions, err = service.NewAuthSessionService(
			repository.NewAuthSessionRepository(b.context.Database),
			b.context.AuthTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize auth sessions: %w", err)
		}
	}

	// Initialize Redis client if not provided
	if b.context.RedisClient == nil {
		redisClient, err := common.GetRedisClient()
		if err != nil {
			common.Logger.Warn("Failed to initialize Redis client", "error", err)
			// Redis is optional, continue without it
			b.context.RedisClient = nil
		} else {
			b.context.RedisClient = redisClient
		}
	}

	// Initialize default services if not provided
	if err := b.initializeStoreIAP(); err != nil {
		return nil, fmt.Errorf("failed to initialize store IAP: %w", err)
	}
	if err := b.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return b.context, nil
}

// initializeServices creates default service instances
func (b *ContextBuilder) initializeServices() error { //nolint:gocyclo // Sequential dependency-injection wiring.
	// Initialize JobService if not provided
	if b.context.JobService == nil {
		// Create minimal River client for job insertion only
		riverClient, err := river.NewClient(riverpgxv5.New(b.context.Database), &river.Config{
			Logger: common.Logger,
		})
		if err != nil {
			return fmt.Errorf("failed to create river client: %w", err)
		}
		b.context.JobService = service.NewJobService(riverClient)
	}

	// Initialize core services if not provided
	if b.context.DefinitionService == nil {
		b.context.DefinitionService = service.NewDefinitionService(b.context.Database)
	}
	if b.context.DefinitionImageService == nil {
		b.context.DefinitionImageService = service.NewDefinitionImageService(b.context.Database)
	}

	// Initialize LeitnerTrackingService if not provided
	if b.context.LeitnerTrackingService == nil {
		b.context.LeitnerTrackingService = service.NewLeitnerTrackingService(b.context.Database)
	}

	if b.context.WordService == nil {
		b.context.WordService = service.NewWordService(b.context.Database, b.context.DefinitionService, b.context.JobService, b.context.LeitnerTrackingService)
	}
	if b.context.WordlistService == nil {
		b.context.WordlistService = service.NewWordlistService(b.context.Database)
	}

	if b.context.RealtimeTelemetryService == nil {
		b.context.RealtimeTelemetryService = service.NewRealtimeTelemetryService(b.context.Database)
	}

	// Initialize MailService early since other services depend on it
	if b.context.MailService == nil {
		b.context.MailService = mail.NewMailService(b.context.Database)
	}

	// Initialize RevenueCatService only while the legacy provider surface is enabled.
	if b.context.RevenueCatService == nil && b.legacyProviderConfig.Enabled {
		if b.revenueCatServiceFactory != nil {
			b.context.RevenueCatService = b.revenueCatServiceFactory(b.context.Database)
		} else {
			apiClient, err := service.NewRevenueCatAPIClient(b.legacyProviderConfig.RevenueCatAPIKey)
			if err != nil {
				return err
			}
			b.context.RevenueCatService = service.NewRevenueCatService(b.context.Database, apiClient)
		}
	}

	if b.context.SubscriptionService == nil {
		b.context.SubscriptionService = service.NewSubscriptionService(
			b.context.Database, b.context.MailService, b.context.RevenueCatService, b.legacyProviderConfig.Stripe,
		)
	}
	if b.context.EffectiveAccessService != nil {
		b.context.SubscriptionService.SetEffectiveAccess(b.context.EffectiveAccessService)
	}

	// Initialize AnalyticsService if not provided
	if b.context.AnalyticsService == nil {
		// Create a base analytics service without user/wordlist specifics for general use
		analyticsService, err := service.NewAnalyticsService(b.context.Database, service.AnalyticsConfig{
			UserID:     0, // Will be set by individual handlers
			WordlistID: 0, // Will be set by individual handlers
			UseCache:   true,
			CacheTTL:   1 * time.Minute,
		})
		if err != nil {
			return fmt.Errorf("failed to create analytics service: %w", err)
		}
		b.context.AnalyticsService = analyticsService
	}

	// Initialize LeitnerSystemStrategy if not provided
	if b.context.LeitnerSystemStrategy == nil {
		// Create analytics writer locally for LeitnerSystemStrategy (non-cached)
		analyticsService, err := service.NewAnalyticsService(b.context.Database, service.AnalyticsConfig{
			UserID:     0,     // Not used for stateless operations
			WordlistID: 0,     // Not used for stateless operations
			UseCache:   false, // No caching for write operations
			CacheTTL:   0,
		})
		if err != nil {
			return fmt.Errorf("failed to create analytics writer: %w", err)
		}

		// Cast to interface pointer
		var analyticsWriter service.LeitnerAnalyticsWriter = analyticsService

		// Create strategy with analytics writer as internal dependency
		b.context.LeitnerSystemStrategy = service.NewLeitnerSystemStrategy(
			b.context.Database,
			b.context.WordService,
			b.context.DefinitionService,
			&analyticsWriter,
			b.context.LeitnerTrackingService,
		)
	}

	// Initialize ErrorReportService if not provided
	if b.context.ErrorReportService == nil {
		b.context.ErrorReportService = service.NewErrorReportService(
			b.context.Database,
			b.context.DefinitionService,
			b.context.WordService,
			b.context.LeitnerTrackingService,
			b.context.JobService,
		)
	}

	// Initialize UserService after subscription/effective-access dependencies.
	if b.context.UserService == nil {
		b.context.UserService = service.NewUserService(
			b.context.Database,
			b.context.SubscriptionService,
			b.context.AuthSessions,
			b.context.JobService,
		)
	}
	if b.context.EffectiveAccessService != nil {
		b.context.UserService.SetEffectiveAccess(b.context.EffectiveAccessService)
	}

	return nil
}

type appleStoreDependencies struct {
	notification *service.AppleNotificationVerifier
	purchase     *service.ApplePurchaseVerifier
	refresher    *service.AppleSubscriptionRefresher
	products     []model.MobileIAPProduct
}

type googleStoreDependencies struct {
	processor *service.GooglePurchaseProcessor
	products  []model.MobileIAPProduct
}

func (b *ContextBuilder) initializeStoreIAP() error {
	storeConfig, err := config.LoadStoreIAPConfig()
	if err != nil {
		return err
	}
	if !storeConfig.Enabled {
		b.context.StoreIAPEnabled = false
		return nil
	}
	if storeConfig.RedisConfigured && b.context.RedisClient == nil {
		return fmt.Errorf("configured Redis limiter is unavailable for store IAP")
	}
	requestLimiter, err := service.NewStoreIAPRequestLimiter(
		b.context.RedisClient, storeConfig.RedisConfigured, nil,
	)
	if err != nil {
		return err
	}
	if !storeConfig.RedisConfigured {
		common.Logger.Warn("store IAP is using degraded process-local rate limiting")
	}
	protector, err := security.NewStoreEvidenceProtector(
		storeConfig.Evidence.LookupKey,
		storeConfig.Evidence.ActiveVersion,
		storeConfig.Evidence.Keys,
	)
	if err != nil {
		return fmt.Errorf("configure protected store evidence: %w", err)
	}
	entitlements, err := repository.NewStoreEntitlementRepository(b.context.Database, protector)
	if err != nil {
		return err
	}
	persistence, err := service.NewStoreEntitlementPersistence(entitlements)
	if err != nil {
		return err
	}
	apple, err := buildAppleStoreDependencies(storeConfig.Apple, storeConfig.ClientEnvironment)
	if err != nil {
		return err
	}
	googleStore, err := buildGoogleStoreDependencies(storeConfig.Google)
	if err != nil {
		return err
	}
	mobileProducts := append(apple.products, googleStore.products...)
	effectiveAccess, err := service.NewEffectiveAccessService(entitlements, storeConfig.ClientEnvironment, nil)
	if err != nil {
		return err
	}
	retryRiver, err := river.NewClient(riverpgxv5.New(b.context.Database), &river.Config{Logger: common.Logger})
	if err != nil {
		return fmt.Errorf("create Google acknowledgement scheduler: %w", err)
	}
	retryScheduler := service.NewJobService(retryRiver)
	acknowledgementWorker, err := service.NewGoogleAcknowledgementRetryWorker(
		entitlements, googleStore.processor, persistence, storeConfig.ClientEnvironment, nil,
	)
	if err != nil {
		return err
	}
	acknowledgementSweep, err := service.NewGoogleAcknowledgementSweepWorker(
		entitlements, acknowledgementWorker, storeConfig.ClientEnvironment,
	)
	if err != nil {
		return err
	}
	orchestrator, err := service.NewStoreIAPOrchestrator(
		entitlements, apple.purchase, apple.refresher, googleStore.processor,
		persistence, retryScheduler, storeConfig.ClientEnvironment, mobileProducts, nil,
	)
	if err != nil {
		return err
	}
	notificationRefresher, err := service.NewStoreNotificationRefresher(
		entitlements, entitlements, apple.refresher, googleStore.processor, persistence,
	)
	if err != nil {
		return err
	}
	inbox := repository.NewProviderEventInboxRepository(b.context.Database)
	appleIngestor, err := service.NewAppleNotificationIngestor(
		apple.notification, inbox, notificationRefresher.RefreshApple, nil,
	)
	if err != nil {
		return err
	}
	googleIngestor, err := service.NewGoogleRTDNIngestor(
		service.GoogleOIDCTokenVerifier{}, inbox, notificationRefresher.RefreshGoogle,
		storeConfig.Google.PushAudience, storeConfig.Google.PushServiceAccountEmail,
		storeConfig.Google.Topic, storeConfig.Google.Subscription,
		storeConfig.Google.PackageName, nil,
	)
	if err != nil {
		return err
	}
	b.context.StoreEntitlementRepository = entitlements
	b.context.ProviderEventInboxRepository = inbox
	b.context.AppleNotificationIngestor = appleIngestor
	b.context.GoogleRTDNIngestor = googleIngestor
	b.context.EffectiveAccessService = effectiveAccess
	b.context.StoreIAPOrchestrator = orchestrator
	b.context.GoogleAcknowledgementWorker = acknowledgementWorker
	b.context.GoogleAcknowledgementSweep = acknowledgementSweep
	b.context.StoreIAPRequestLimiter = requestLimiter
	b.context.StoreIAPEnabled = true
	return nil
}

func buildAppleStoreDependencies(
	storeConfig config.AppleStoreConfig,
	environment model.StoreEnvironment,
) (appleStoreDependencies, error) {
	jws, err := service.NewAppleJWSVerifier(storeConfig.Roots)
	if err != nil {
		return appleStoreDependencies{}, fmt.Errorf("configure Apple signed-data verifier: %w", err)
	}
	client, err := service.NewAppleAppStoreClient(service.AppleAppStoreClientConfig{
		KeyID: storeConfig.KeyID, IssuerID: storeConfig.IssuerID,
		BundleID: storeConfig.BundleID, PrivateKey: storeConfig.PrivateKey,
	})
	if err != nil {
		return appleStoreDependencies{}, fmt.Errorf("configure Apple App Store client: %w", err)
	}
	providerProducts := make([]string, 0, len(storeConfig.Products))
	allowlist := make(map[string]string, len(storeConfig.Products))
	mobileProducts := make([]model.MobileIAPProduct, 0, len(storeConfig.Products))
	for productID, product := range storeConfig.Products {
		providerProducts = append(providerProducts, productID)
		allowlist[productID] = product.Entitlement
		mobileProducts = append(mobileProducts, model.MobileIAPProduct{
			Store: model.EntitlementStoreApple, ProductID: productID,
			Entitlement: product.Entitlement, BillingPeriod: product.BillingPeriod,
		})
	}
	notification, err := service.NewAppleNotificationVerifier(
		jws, storeConfig.BundleID, storeConfig.AppAppleID, providerProducts, nil,
	)
	if err != nil {
		return appleStoreDependencies{}, err
	}
	refresher, err := service.NewAppleSubscriptionRefresher(
		client, jws, storeConfig.BundleID, storeConfig.AppAppleID, allowlist, nil,
	)
	if err != nil {
		return appleStoreDependencies{}, err
	}
	purchase, err := service.NewApplePurchaseVerifier(
		client, jws, storeConfig.BundleID, allowlist, environment, nil,
	)
	if err != nil {
		return appleStoreDependencies{}, err
	}
	return appleStoreDependencies{
		notification: notification, purchase: purchase, refresher: refresher, products: mobileProducts,
	}, nil
}

func buildGoogleStoreDependencies(storeConfig config.GooglePlayConfig) (googleStoreDependencies, error) {
	credentialsContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	credentials, err := google.FindDefaultCredentials(credentialsContext, googleAndroidPublisherScope)
	if err != nil {
		return googleStoreDependencies{}, fmt.Errorf("find Google Application Default Credentials: %w", err)
	}
	tokens, err := service.NewGoogleOAuthTokenSource(credentials.TokenSource)
	if err != nil {
		return googleStoreDependencies{}, err
	}
	client, err := service.NewGooglePlayClient(storeConfig.PackageName, tokens, nil, "")
	if err != nil {
		return googleStoreDependencies{}, err
	}
	allowlist := make(map[string]string, len(storeConfig.Products))
	mobileProducts := make([]model.MobileIAPProduct, 0, len(storeConfig.Products))
	for productID, product := range storeConfig.Products {
		allowlist[productID] = product.Entitlement
		mobileProducts = append(mobileProducts, model.MobileIAPProduct{
			Store: model.EntitlementStoreGoogle, ProductID: productID,
			Entitlement: product.Entitlement, BillingPeriod: product.BillingPeriod,
		})
	}
	verifier, err := service.NewGooglePurchaseVerifier(client, allowlist, nil)
	if err != nil {
		return googleStoreDependencies{}, err
	}
	processor, err := service.NewGooglePurchaseProcessor(verifier, client)
	if err != nil {
		return googleStoreDependencies{}, err
	}
	return googleStoreDependencies{processor: processor, products: mobileProducts}, nil
}

// Close gracefully shuts down all services and connections
func (ctx *Context) Close() {
	log.Println("Shutting down Context...")

	// Close database connection
	if ctx.Database != nil {
		ctx.Database.Close()
		log.Println("Database connection closed")
	}

	// Close Redis connection
	if ctx.RedisClient != nil {
		if err := ctx.RedisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}

	log.Println("Context shutdown complete")
}

// GetDatabase returns the database connection
func (ctx *Context) GetDatabase() *pgxpool.Pool {
	return ctx.Database
}

// GetJobService returns the JobService
func (ctx *Context) GetJobService() service.JobService {
	return ctx.JobService
}
