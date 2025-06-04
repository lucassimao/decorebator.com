# Dependency Injection Modernization Plan

## 🔍 Current Architecture Analysis

### **Major Anti-Patterns Found:**

1. **Global variables with `init()` functions** in service layer
2. **Mixed dependency injection approaches** (some proper DI, mostly globals)
3. **No interface abstractions** for repositories/services
4. **`os.Exit(1)` calls in `init()`** making testing impossible
5. **Manual dependency wiring** scattered across the codebase

### **Good Patterns Found:**
- Database singleton with `sync.Once` (well-implemented)
- Some proper constructor DI in subscription service
- Centralized environment configuration

### **Current Problematic Patterns:**

#### Service Layer Anti-Patterns:
```go
// internal/service/word.go
var wordRepository *repo.WordRepository
func init() {
    db, err := common.GetDBConnection()
    if err != nil {
        os.Exit(1) // Kills testing
    }
    wordRepository = &repo.WordRepository{Db: db}
}

// internal/service/user.go  
var userRepository *repo.UserRepository
func init() {
    db, err := common.GetDBConnection()
    if err != nil {
        os.Exit(1) // Kills testing
    }
    userRepository = &repo.UserRepository{Db: db}
}
```

#### Mixed DI in HTTP Layer:
```go
// internal/http/setup.go
// Some services get proper DI
subService := service.NewSubscriptionService(db)

// Others use globals
service.SaveUser(firstName, lastName, password, email)
```

## 📋 Actionable Modernization Plan

### **Phase 1: Interface Abstraction (High Priority)**

#### 1.1 Create Repository Interfaces
**Files to create:**
- `internal/repository/interfaces.go`

**Define interfaces for all repositories:**
```go
type UserRepository interface {
    Save(firstName, lastName, password, email string) (*model.User, error)
    Find(args FindUserArgs) ([]model.User, error)
    GetByID(id int64) (*model.User, error)
    GetByEmail(email string) (*model.User, error)
    UpdatePassword(userID int64, hashedPassword string) error
    Delete(id int64) error
}

type WordRepository interface {
    Save(word, notes string, userID, wordlistID int64, tx *pgx.Tx) (*model.Word, error)
    GetWordsByWordlist(wordlistID int64) ([]model.Word, error)
    GetByID(id int64) (*model.Word, error)
    Update(id int64, word, notes string) error
    Delete(id int64) error
    GetRandomWordsByWordlist(wordlistID int64, limit int) ([]model.Word, error)
}

type WordlistRepository interface {
    Save(name, language string, userID int64) (*model.Wordlist, error)
    GetByUserID(userID int64) ([]model.Wordlist, error)
    GetByID(id int64) (*model.Wordlist, error)
    Update(id int64, name, language string) error
    Delete(id int64) error
    GetWordCount(wordlistID int64) (int, error)
}

type DefinitionRepository interface {
    Save(definition *model.Definition, tokenID string) (*model.Definition, error)
    GetByWordID(wordID int64) ([]model.Definition, error)
    GetByID(id int64) (*model.Definition, error)
    Update(definition *model.Definition) error
    Delete(id int64) error
}

type SubscriptionRepository interface {
    Save(subscription *model.Subscription) (*model.Subscription, error)
    GetByUserID(userID int64) (*model.Subscription, error)
    GetByID(id int64) (*model.Subscription, error)
    Update(subscription *model.Subscription) error
    GetActiveSubscription(userID int64) (*model.Subscription, error)
}
```

#### 1.2 Create Service Interfaces
**Files to create:**
- `internal/service/interfaces.go`

**Define service contracts:**
```go
type UserService interface {
    SaveUser(firstName, lastName, password, email string) (*model.User, error)
    AuthenticateUser(email, password string) (*model.User, error)
    GetUserByID(id int64) (*model.User, error)
    GetUserByEmail(email string) (*model.User, error)
    UpdatePassword(userID int64, newPassword string) error
    DeleteUser(id int64) error
    SendPasswordResetEmail(email string) error
}

type WordService interface {
    CreateWord(input CreateWordInput) (*model.Word, error)
    GetWordsByWordlist(wordlistID int64) ([]model.Word, error)
    GetWordByID(id int64) (*model.Word, error)
    UpdateWord(id int64, word, notes string) error
    DeleteWord(id int64) error
    GetRandomWordsForQuiz(wordlistID int64, limit int) ([]model.Word, error)
}

type WordlistService interface {
    CreateWordlist(name, language string, userID int64) (*model.Wordlist, error)
    GetWordlistsByUserID(userID int64) ([]model.Wordlist, error)
    GetWordlistByID(id int64) (*model.Wordlist, error)
    UpdateWordlist(id int64, name, language string) error
    DeleteWordlist(id int64) error
    GetWordlistWithStats(id int64) (*model.WordlistWithStats, error)
}

type DefinitionService interface {
    CreateDefinition(definition *model.Definition, tokenID string) (*model.Definition, error)
    GetDefinitionsByWordID(wordID int64) ([]model.Definition, error)
    FetchAndSaveDefinitions(wordID int64) error
    UpdateDefinition(definition *model.Definition) error
}

type SubscriptionService interface {
    CreateCheckoutSession(userID int64, priceID string) (string, error)
    GetSubscriptionByUserID(userID int64) (*model.Subscription, error)
    UpdateSubscriptionFromWebhook(event StripeEvent) error
    CancelSubscription(userID int64) error
    IsUserPremium(userID int64) (bool, error)
}
```

### **Phase 2: Remove Global State (Critical Priority)**

#### 2.1 Eliminate Service Layer Globals
**Files to modify:**

**`internal/service/word.go`:**
- Remove `var wordRepository *repo.WordRepository`
- Remove `init()` function
- Convert all functions to methods on a struct
- Add constructor: `NewWordService(repo repository.WordRepository) WordService`

**Target transformation:**
```go
// FROM:
var wordRepository *repo.WordRepository

func init() {
    db, err := common.GetDBConnection()
    if err != nil {
        os.Exit(1)
    }
    wordRepository = &repo.WordRepository{Db: db}
}

func CreateWord(input CreateWordInput) (*model.Word, error) {
    return wordRepository.Save(input.Word, input.Notes, input.UserID, input.WordlistID, nil)
}

// TO:
type wordService struct {
    wordRepo repository.WordRepository
    logger   *slog.Logger
}

func NewWordService(wordRepo repository.WordRepository) WordService {
    return &wordService{
        wordRepo: wordRepo,
        logger:   common.Logger.With("service", "word"),
    }
}

func (s *wordService) CreateWord(input CreateWordInput) (*model.Word, error) {
    s.logger.Info("creating word", "wordName", input.Word, "userID", input.UserID)
    return s.wordRepo.Save(input.Word, input.Notes, input.UserID, input.WordlistID, nil)
}
```

**`internal/service/user.go`:**
```go
// FROM:
var userRepository *repo.UserRepository

func SaveUser(firstName, lastName, password, email string) (*model.User, error) {
    // implementation
}

// TO:
type userService struct {
    userRepo repository.UserRepository
    logger   *slog.Logger
}

func NewUserService(userRepo repository.UserRepository) UserService {
    return &userService{
        userRepo: userRepo,
        logger:   common.Logger.With("service", "user"),
    }
}

func (s *userService) SaveUser(firstName, lastName, password, email string) (*model.User, error) {
    s.logger.Info("saving user", "email", email)
    // implementation
}
```

**`internal/service/wordlist.go`:**
- Same pattern as above
- Remove global repository variables
- Add proper constructor injection

#### 2.2 Convert to Constructor-Based DI
**Target pattern for all services:**
```go
type serviceStruct struct {
    repo   repository.RepositoryInterface
    logger *slog.Logger
    // other dependencies
}

func NewService(repo repository.RepositoryInterface) ServiceInterface {
    return &serviceStruct{
        repo:   repo,
        logger: common.Logger.With("service", "serviceName"),
    }
}
```

### **Phase 3: Dependency Container (Medium Priority)**

#### 3.1 Create Service Container
**File to create:** `internal/container/container.go`

**Implement dependency container:**
```go
package container

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "decorebator/internal/repository"
    "decorebator/internal/service"
    "decorebator/internal/common"
)

type Container struct {
    // Database
    db *pgxpool.Pool
    
    // Repositories
    userRepo         repository.UserRepository
    wordRepo         repository.WordRepository
    wordlistRepo     repository.WordlistRepository
    definitionRepo   repository.DefinitionRepository
    subscriptionRepo repository.SubscriptionRepository
    
    // Services
    userService         service.UserService
    wordService         service.WordService
    wordlistService     service.WordlistService
    definitionService   service.DefinitionService
    subscriptionService service.SubscriptionService
}

func NewContainer() (*Container, error) {
    // Initialize database
    db, err := common.GetDBConnection()
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Create repositories
    userRepo := repository.NewUserRepository(db)
    wordRepo := repository.NewWordRepository(db)
    wordlistRepo := repository.NewWordlistRepository(db)
    definitionRepo := repository.NewDefinitionRepository(db)
    subscriptionRepo := repository.NewSubscriptionRepository(db)
    
    // Create services with injected repositories
    userService := service.NewUserService(userRepo)
    wordService := service.NewWordService(wordRepo)
    wordlistService := service.NewWordlistService(wordlistRepo)
    definitionService := service.NewDefinitionService(definitionRepo)
    subscriptionService := service.NewSubscriptionService(subscriptionRepo)
    
    return &Container{
        db:                  db,
        userRepo:           userRepo,
        wordRepo:           wordRepo,
        wordlistRepo:       wordlistRepo,
        definitionRepo:     definitionRepo,
        subscriptionRepo:   subscriptionRepo,
        userService:        userService,
        wordService:        wordService,
        wordlistService:    wordlistService,
        definitionService:  definitionService,
        subscriptionService: subscriptionService,
    }, nil
}

// Service getters
func (c *Container) UserService() service.UserService { return c.userService }
func (c *Container) WordService() service.WordService { return c.wordService }
func (c *Container) WordlistService() service.WordlistService { return c.wordlistService }
func (c *Container) DefinitionService() service.DefinitionService { return c.definitionService }
func (c *Container) SubscriptionService() service.SubscriptionService { return c.subscriptionService }

// Repository getters (for advanced use cases)
func (c *Container) UserRepository() repository.UserRepository { return c.userRepo }
func (c *Container) WordRepository() repository.WordRepository { return c.wordRepo }
func (c *Container) WordlistRepository() repository.WordlistRepository { return c.wordlistRepo }

// Database getter
func (c *Container) Database() *pgxpool.Pool { return c.db }

// Cleanup
func (c *Container) Close() {
    common.CloseDBConnection()
}
```

### **Phase 4: HTTP Layer Refactoring (Medium Priority)**

#### 4.1 Controller Dependency Injection
**Files to modify:**
- `internal/http/setup.go` - Refactor to use container

**New setup pattern:**
```go
func SetupRoutes() *gin.Engine {
    container, err := container.NewContainer()
    if err != nil {
        log.Fatal("Failed to initialize container:", err)
    }
    
    router := gin.Default()
    
    // Add middleware
    router.Use(cors.Default())
    
    // Inject services into route handlers
    setupUserRoutes(router, container.UserService())
    setupWordRoutes(router, container.WordService())
    setupWordlistRoutes(router, container.WordlistService())
    setupSubscriptionRoutes(router, container.SubscriptionService())
    
    return router
}
```

#### 4.2 Handler Function Signatures
**Create separate handler files:**
- `internal/http/handlers/user_handlers.go`
- `internal/http/handlers/word_handlers.go`
- `internal/http/handlers/wordlist_handlers.go`
- `internal/http/handlers/subscription_handlers.go`

**Convert from:**
```go
router.POST("/words", func(c *gin.Context) {
    // Uses global service.CreateWord()
})
```

**To:**
```go
// internal/http/handlers/word_handlers.go
func SetupWordRoutes(router *gin.Engine, wordService service.WordService) {
    router.POST("/words", createWordHandler(wordService))
    router.GET("/wordlists/:id/words", getWordsHandler(wordService))
    router.PUT("/words/:id", updateWordHandler(wordService))
    router.DELETE("/words/:id", deleteWordHandler(wordService))
}

func createWordHandler(wordService service.WordService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input service.CreateWordInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        word, err := wordService.CreateWord(input)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create word"})
            return
        }
        
        c.JSON(http.StatusCreated, word)
    }
}
```

### **Phase 5: Repository Implementation Updates (Medium Priority)**

#### 5.1 Implement Repository Interfaces
**Files to modify:**
- `internal/repository/user.go` - Ensure implements UserRepository interface
- `internal/repository/word.go` - Ensure implements WordRepository interface
- `internal/repository/wordlist.go` - Ensure implements WordlistRepository interface

#### 5.2 Constructor Standardization
**Ensure all repositories have:**
```go
// internal/repository/user.go
type userRepository struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
    return &userRepository{db: db}
}

// Verify interface compliance at compile time
var _ UserRepository = (*userRepository)(nil)
```

### **Phase 6: Database Connection Refactoring (Low Priority)**

#### 6.1 Remove os.Exit from Database Connection
**File to modify:** `internal/common/database.go`

**Change from:**
```go
func GetDBConnection() (*pgxpool.Pool, error) {
    dbOnce.Do(func() {
        var err error
        db, err = pgxpool.New(context.Background(), Env.DatabaseUrl)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
            os.Exit(1) // REMOVE THIS
        }
    })
    return db, nil
}
```

**To:**
```go
func GetDBConnection() (*pgxpool.Pool, error) {
    var initErr error
    dbOnce.Do(func() {
        var err error
        db, err = pgxpool.New(context.Background(), Env.DatabaseUrl)
        if err != nil {
            initErr = fmt.Errorf("unable to connect to database: %w", err)
            return
        }
    })
    
    if initErr != nil {
        return nil, initErr
    }
    
    return db, nil
}
```

## 🎯 Implementation Priority

### **Critical (This Week)**
1. Remove `init()` functions with `os.Exit(1)` calls
2. Create repository interfaces (`internal/repository/interfaces.go`)
3. Convert one service (start with `word.go`) to constructor-based DI

### **High Priority (Next 2 Weeks)**
1. Convert all remaining services to constructor-based DI
2. Create service interfaces (`internal/service/interfaces.go`)
3. Implement basic dependency container (`internal/container/container.go`)

### **Medium Priority (Next Month)**
1. Refactor HTTP layer to use dependency injection
2. Create separate handler files with proper DI
3. Standardize all repository constructors

## 🔧 Specific Files to Modify

### **Immediate Changes Needed:**
- `internal/service/word.go` - Remove global `wordRepository`, add constructor
- `internal/service/user.go` - Remove global `userRepository`, add constructor  
- `internal/service/wordlist.go` - Remove global `wordlistRepository`, add constructor
- `internal/common/database.go` - Remove `os.Exit(1)` calls

### **Architecture Files to Create:**
- `internal/repository/interfaces.go` - Repository contracts
- `internal/service/interfaces.go` - Service contracts
- `internal/container/container.go` - Dependency container
- `internal/http/handlers/` - Separate handler functions with DI

### **Testing Improvements After Modernization:**
After these changes, you'll be able to:
- ✅ Mock repository interfaces for unit testing
- ✅ Test services in isolation without database
- ✅ Inject test databases for integration tests
- ✅ Avoid global state in tests
- ✅ Run tests in parallel safely
- ✅ Create focused unit tests for business logic

### **Example Test After Modernization:**
```go
func TestWordService_CreateWord(t *testing.T) {
    // Setup mock repository
    mockRepo := &mocks.WordRepository{}
    wordService := service.NewWordService(mockRepo)
    
    // Setup expectations
    mockRepo.On("Save", "hello", "greeting", int64(1), int64(1), nil).
        Return(&model.Word{ID: 1, Name: "hello"}, nil)
    
    // Test
    word, err := wordService.CreateWord(service.CreateWordInput{
        Word: "hello",
        Notes: "greeting",
        UserID: 1,
        WordlistID: 1,
    })
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, "hello", word.Name)
    mockRepo.AssertExpectations(t)
}
```

This modernization will make the codebase more testable, maintainable, and follow Go best practices for dependency management.