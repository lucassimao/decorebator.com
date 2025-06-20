# Worker Abuse Prevention Implementation

## Overview

This document describes the implementation of abuse prevention for background workers to prevent free users from generating excessive AI-powered content beyond their tier limits.

## Free Tier Limitations

Free users are restricted to:
- **1 wordlist maximum**
- **10 words total** across all wordlists

These limits are enforced at multiple levels:
1. **API level** - When creating new words
2. **Worker level** - Before processing any AI generation tasks
3. **Error reporting level** - Before regenerating content

## Implementation Details

### 1. Validation Functions (`worker_validation.go`)

#### `ValidateUserEligibilityForWorkers(userID int64) error`
- Checks if a user is eligible for worker processing
- Premium users (monthly/annual) have no restrictions
- Free users are validated against wordlist and word count limits
- Returns a `BusinessError` with upgrade prompt if limits exceeded

#### `ValidateWordEligibilityForWorkers(wordID int64) error`
- Validates eligibility for word-based workers (definition fetching, text-to-speech)
- Looks up user from word ID and calls main validation function

#### `ValidateDefinitionEligibilityForWorkers(definitionID int64) error`
- Validates eligibility for definition-based workers (image generation)
- Looks up user from definition → word relationship
- Allows shared definitions (when no user found)

### 2. Worker Integration

All workers now include validation at the start of their `Work()` method:

#### Definition Fetcher Worker
```go
if err := ValidateWordEligibilityForWorkers(job.Args.WordId); err != nil {
    return river.JobCancel(err) // Permanent cancellation
}
```

#### Image Generator Worker
```go
if err := ValidateDefinitionEligibilityForWorkers(definitionID); err != nil {
    return river.JobCancel(err)
}
```

#### Text-to-Speech Worker
```go
if err := ValidateWordEligibilityForWorkers(job.Args.WordId); err != nil {
    return river.JobCancel(err)
}
```

#### Example Audio Worker
```go
if err := ValidateWordEligibilityForWorkers(job.Args.WordID); err != nil {
    return river.JobCancel(err)
}
```

### 3. Prevention Points

#### Word Creation (`word.go`)
- Validation occurs before saving word to database
- Prevents creation if user exceeds limits
- Provides immediate feedback to user

#### Error Reporting (`error_reporting.go`)
- Validation before triggering regeneration workers
- Prevents abuse through error reporting system
- Maintains rate limiting for additional protection

## Error Handling

When limits are exceeded:
1. **Job Cancellation**: Workers use `river.JobCancel(err)` for permanent failures
2. **User Feedback**: Clear error messages indicating upgrade requirement
3. **Logging**: Warn-level logs for monitoring abuse patterns

## Security Considerations

1. **Defense in Depth**: Validation at multiple layers (API, worker, error reporting)
2. **Permanent Cancellation**: Jobs don't retry when user is ineligible
3. **Resource Protection**: Prevents costly AI API calls for free users beyond limits
4. **Upgrade Path**: Clear messaging directs users to subscription options

## Testing

Comprehensive test suite in `worker_validation_test.go` covers:
- Premium user validation (no restrictions)
- Free user validation (various scenarios)
- Boundary conditions (exactly at limits)
- Exceeded limits (wordlists and words)
- Word-specific validation

## Configuration

Limits are defined in `model/subscription.go`:
```go
const (
    FreeWordlistLimit = 1
    FreeWordsPerList  = 10
)
```

## Monitoring

All validation failures are logged with:
- User ID
- Resource being accessed
- Validation failure reason
- Timestamp for analysis

This enables monitoring of abuse patterns and system usage analytics.