# Feature Tickets: Vocabulary Retention Improvements

These tickets address the **production gap** in vocabulary learning - users can recognize words during quizzes but struggle to recall them when needed in real contexts (emails, Slack messages, etc.).

---

## Ticket 1: Reverse Search by Meaning (Find Word By Definition)

### Summary
Add the ability to search across all wordlists by definition/meaning, not just by word name. This solves the "I know I have a word for this but can't remember it" problem.

### Labels
`enhancement` `high-priority` `api` `mobile`

### Problem Statement
Users frequently know they've learned a word that fits a specific context but cannot recall it. The current search:
- Only searches word names (not definitions)
- Is client-side only in mobile app
- Is scoped to a single wordlist
- Uses simple substring matching

When writing an email or Slack message, users need to search by **what they want to say**, not by the word they can't remember.

### User Story
> As a user, I want to search "politely decline" and find my word "demur" so that I can use vocabulary I've learned in real situations.

### Proposed Solution

**API Changes:**

1. Add new endpoint: `GET /words/search?q=<query>`
   - Searches across ALL user's wordlists
   - Searches in: `definitions.meaning`, `definitions.examples`, `words.name`
   - Returns words with matching definitions, grouped by wordlist

2. Use PostgreSQL full-text search for performance:
   ```sql
   -- Add tsvector column to definitions table
   ALTER TABLE definitions ADD COLUMN search_vector tsvector;

   -- Create GIN index
   CREATE INDEX idx_definitions_search ON definitions USING GIN(search_vector);

   -- Populate with meaning and examples
   UPDATE definitions SET search_vector =
     to_tsvector('english', coalesce(meaning, '') || ' ' || coalesce(examples, ''));
   ```

3. Query example:
   ```sql
   SELECT w.name, d.meaning, wl.name as wordlist_name
   FROM definitions d
   JOIN word_definitions wd ON wd.definition_id = d.id
   JOIN words w ON w.id = wd.word_id
   JOIN wordlists wl ON wl.id = w.wordlist_id
   WHERE d.search_vector @@ plainto_tsquery('english', $1)
     AND wl.user_id = $2
   ORDER BY ts_rank(d.search_vector, plainto_tsquery('english', $1)) DESC
   LIMIT 20;
   ```

**Mobile Changes:**

1. Add global search screen accessible from dashboard
2. Search input with debounced API calls (300ms)
3. Results grouped by wordlist with word name, definition preview
4. Tap result to view full definition or copy word

**File Locations:**
- API handler: `api/internal/http/word.go` (add new handler)
- Repository: `api/internal/repository/word.go` (add search method)
- Migration: `api/migrations/` (add tsvector column and index)
- Mobile: `mobile/app/search.tsx` (new screen)
- Mobile API: `mobile/api/words.ts` (add search function)

### Acceptance Criteria
- [ ] User can search by partial meaning across all wordlists
- [ ] Search returns results ranked by relevance
- [ ] Results show word, definition snippet, and source wordlist
- [ ] Search is performant (<500ms for typical queries)
- [ ] Mobile UI shows results with tap-to-copy functionality

### Effort Estimate
Medium (3-5 days)

---

## Ticket 2: Introduce Production Quizzes Earlier in Learning Cycle

### Summary
Move `WriteWordFromDefinition` quiz type from boxes 5-7 to box 3 to force active recall earlier in the learning process, strengthening long-term retention.

### Labels
`enhancement` `high-priority` `api` `learning-algorithm`

### Problem Statement
The current Leitner box-to-quiz-type mapping prioritizes recognition in early boxes:

| Box | Current Quiz Types | Memory Mode |
|-----|-------------------|-------------|
| 1 | GuessMeaning | Recognition only |
| 2 | WordFromMeaning, GuessMeaning | Recognition |
| 3 | WordFromImage, WordFromMeaning | Recognition |
| 4 | CompleteSentence, WordFromExampleAudio, GuessMeaning | Contextual |
| 5+ | WriteWordFromDefinition... | **Production** |

Research shows that **active recall** (producing the word) builds stronger neural pathways than passive recognition (selecting from options). By the time users reach box 5, they've already reinforced recognition patterns without production practice.

### User Story
> As a learner, I want to practice typing/producing words earlier so that I can recall them when writing, not just recognize them when reading.

### Proposed Solution

**Update quiz type mapping in** `api/internal/service/leitner_system_strategy.go`:

```go
// Current (lines 93-101)
var boxQuizTypes = map[int][]model.QuizType{
    1: {model.GuessMeaning},
    2: {model.WordFromMeaning, model.GuessMeaning},
    3: {model.WordFromImage, model.WordFromMeaning},
    // ...
}

// Proposed
var boxQuizTypes = map[int][]model.QuizType{
    1: {model.GuessMeaning},
    2: {model.WordFromMeaning, model.GuessMeaning},
    3: {model.WordFromImage, model.WordFromMeaning, model.WriteWordFromDefinition}, // Added production
    4: {model.CompleteSentence, model.WordFromExampleAudio, model.WriteWordFromDefinition}, // Added production
    // Box 5+ remains the same
}
```

**Rationale for Box 3:**
- User has seen the word at least twice (box 1 → 2 → 3)
- 24+ hours have passed, testing true retention
- Early enough to correct weak encoding before it solidifies
- Image quiz provides visual anchor, then production tests active recall

### File Locations
- `api/internal/service/leitner_system_strategy.go` lines 93-101

### Acceptance Criteria
- [ ] `WriteWordFromDefinition` appears in box 3 quiz rotation
- [ ] `WriteWordFromDefinition` appears in box 4 quiz rotation
- [ ] Quiz type balancing algorithm includes new types in selection pool
- [ ] Existing tests pass with updated mapping
- [ ] Add test case verifying production quiz appears in box 3

### Risks & Mitigations
- **Risk:** Users may find it too hard early on
  - **Mitigation:** Keep it as ONE option among several; balancing algorithm will rotate through types
- **Risk:** May increase incorrect answers in early boxes
  - **Mitigation:** Monitor analytics; can adjust if retention metrics decline

### Effort Estimate
Low (1-2 hours) - Configuration change only

---

## Ticket 3: Sentence Generation Quiz Type

### Summary
Add a new quiz type where users must write a sentence using the target word, forcing production in context rather than isolated recall.

### Labels
`enhancement` `medium-priority` `api` `mobile` `ai-integration`

### Problem Statement
Current production quiz (`WriteWordFromDefinition`) tests isolated word recall. But real-world usage requires:
1. Recalling the word
2. Understanding its grammatical role
3. Constructing appropriate context

Users can type "obviate" when shown its definition but struggle to use it naturally in an email.

### User Story
> As a learner, I want to practice using words in sentences so that I can confidently use them in real writing.

### Proposed Solution

**New Quiz Type: `UseInSentence`**

**Quiz Flow:**
```
┌─────────────────────────────────────────────────┐
│  ✍️ Use It                                      │
│                                                 │
│  Word: obviate                                  │
│  Meaning: to remove or prevent (a difficulty)  │
│                                                 │
│  Write a sentence using this word:             │
│  ┌─────────────────────────────────────────┐   │
│  │                                         │   │
│  └─────────────────────────────────────────┘   │
│                                                 │
│  Context hint: Work / Professional             │
│                                                 │
│  [Submit →]                                     │
└─────────────────────────────────────────────────┘
```

**Two Implementation Options:**

**Option A: AI-Validated (Recommended for Premium)**
1. User submits sentence
2. Send to OpenAI for validation:
   ```
   Prompt: "Does this sentence correctly use the word '{word}'
   meaning '{definition}'? Sentence: '{user_sentence}'

   Respond with JSON: {
     "correct": boolean,
     "feedback": "brief explanation",
     "suggestion": "improved version if incorrect"
   }"
   ```
3. Show feedback and mark correct/incorrect

**Option B: Self-Reported (Free tier)**
1. User writes sentence
2. Show model example from definition's `examples` field
3. User self-reports: "I used it correctly" / "I need to review"
4. Honor system with example comparison

**API Changes:**

1. Add quiz type to `api/internal/model/quiz.go`:
   ```go
   const (
       // ... existing types
       UseInSentence QuizType = "use_in_sentence"
   )
   ```

2. Add to box mapping (boxes 4-7) in `leitner_system_strategy.go`

3. Create validation endpoint: `POST /quiz/validate-sentence`
   ```go
   type ValidateSentenceRequest struct {
       Word      string `json:"word"`
       Meaning   string `json:"meaning"`
       Sentence  string `json:"sentence"`
   }

   type ValidateSentenceResponse struct {
       Correct    bool   `json:"correct"`
       Feedback   string `json:"feedback"`
       Suggestion string `json:"suggestion,omitempty"`
   }
   ```

4. Add OpenAI integration in `api/internal/openai/` for sentence validation

**Mobile Changes:**

1. New quiz card component: `mobile/components/quiz/UseInSentenceCard.tsx`
2. Text input with word highlighted/locked
3. Context hint selector (casual, professional, academic)
4. Feedback display with animation

**File Locations:**
- Model: `api/internal/model/quiz.go`
- Quiz logic: `api/internal/service/leitner_system_strategy.go`
- New endpoint: `api/internal/http/quiz.go`
- OpenAI: `api/internal/openai/sentence_validator.go` (new file)
- Mobile card: `mobile/components/quiz/UseInSentenceCard.tsx` (new file)

### Acceptance Criteria
- [ ] New `UseInSentence` quiz type defined in model
- [ ] Quiz appears in box 4+ rotation
- [ ] User can type sentence with word
- [ ] AI validates sentence correctness (premium)
- [ ] Self-report fallback works (free)
- [ ] Feedback shows after submission
- [ ] Analytics track this quiz type

### Context Hints (Optional Enhancement)
Provide situational prompts to make practice more realistic:
- "Use in a work email"
- "Use in casual conversation"
- "Use in a formal report"

### Effort Estimate
Medium-High (5-7 days)

---

## Ticket 4: Memory Palace Prompts for Enhanced Retention

### Summary
Integrate Memory Palace (Method of Loci) technique by prompting users to create spatial associations when learning new words, then testing recall via location cues.

### Labels
`enhancement` `medium-priority` `api` `mobile` `experimental`

### Problem Statement
Words are learned as isolated facts with single retrieval paths (definition → word). Research shows Memory Palace technique achieves **95.2% recall** vs **81.9% for flashcards alone** by creating spatial anchors.

When users can't recall a word, they have no alternative retrieval path. Memory Palace creates a second path: location → visual scene → word.

### User Story
> As a learner, I want to associate words with mental locations so that I have multiple ways to recall them when needed.

### Proposed Solution

**Phase 1: Memory Hints in Definitions**

Add AI-generated memory hints that suggest vivid imagery:

1. Extend definition generation prompt to include memory hint:
   ```
   For the word "{word}" meaning "{meaning}", suggest a memorable
   visual scene or mnemonic. Use wordplay, vivid imagery, or
   sound-alike associations. Keep it brief (1-2 sentences).

   Example for "perspicacious" (keen perception):
   "Picture someone wearing SPECTACLES (per-SPIC-acious) reading
   impossibly tiny text with ease."
   ```

2. Add `memory_hint` column to definitions table:
   ```sql
   ALTER TABLE definitions ADD COLUMN memory_hint TEXT;
   ```

3. Display memory hint during first encounter with word

**Phase 2: Location Association (User-Driven)**

1. Add user location preferences:
   ```sql
   CREATE TABLE user_memory_locations (
       id SERIAL PRIMARY KEY,
       user_id INTEGER REFERENCES users(id),
       name VARCHAR(50) NOT NULL,  -- "Kitchen", "Office", "Commute"
       created_at TIMESTAMP DEFAULT NOW()
   );

   ALTER TABLE user_word_progress ADD COLUMN memory_location_id INTEGER
       REFERENCES user_memory_locations(id);
   ```

2. During learning, prompt user to place word in a location:
   ```
   "Where will you remember this? [Kitchen ▼] [Office] [+ Add place]"
   ```

3. Store association for later recall quiz

**Phase 3: Location-Based Recall Quiz**

New quiz type: `RecallFromLocation`

```
┌─────────────────────────────────────────────────┐
│  🧠 Memory Walk                                 │
│                                                 │
│  Walk to your KITCHEN in your mind...           │
│                                                 │
│  You see someone wearing spectacles,            │
│  reading impossibly tiny text.                  │
│                                                 │
│  What word is this?                             │
│  ┌─────────────────────────────────────────┐   │
│  │                                         │   │
│  └─────────────────────────────────────────┘   │
│                                                 │
│  [Submit]                                       │
└─────────────────────────────────────────────────┘
```

**Implementation Phases:**

| Phase | Scope | Effort |
|-------|-------|--------|
| 1 | Memory hints in definitions | Low (2-3 days) |
| 2 | Location associations | Medium (3-4 days) |
| 3 | Location recall quiz | Medium (3-4 days) |

### File Locations
- Definition model: `api/internal/model/definition.go`
- OpenAI prompts: `api/internal/openai/definition.go`
- Migration: `api/migrations/`
- Quiz type: `api/internal/model/quiz.go`
- Mobile UI: `mobile/components/quiz/MemoryPalaceCard.tsx` (new)
- Mobile onboarding: `mobile/app/onboarding/` (add location setup)

### Acceptance Criteria

**Phase 1:**
- [ ] Memory hint field added to definitions
- [ ] OpenAI prompt generates memory hints for new words
- [ ] Backfill job creates hints for existing definitions
- [ ] Mobile displays memory hint on first word encounter

**Phase 2:**
- [ ] User can create custom memory locations
- [ ] User can associate word with location during learning
- [ ] Association stored in database

**Phase 3:**
- [ ] `RecallFromLocation` quiz type implemented
- [ ] Quiz shows location + memory hint, user types word
- [ ] Appears in box 5+ rotation (advanced recall)

### Research References
- Method of Loci neural mechanisms: https://pmc.ncbi.nlm.nih.gov/articles/PMC7929507/
- Memory Palace vs Flashcards study: 95.2% vs 81.9% recall

### Risks & Mitigations
- **Risk:** Users may not engage with location prompts
  - **Mitigation:** Make it optional; track engagement metrics
- **Risk:** AI-generated hints may be low quality
  - **Mitigation:** Allow user editing; add report button

### Effort Estimate
Medium-High (8-10 days total across all phases)

---

## Implementation Priority

| Ticket | Impact | Effort | Recommended Order |
|--------|--------|--------|-------------------|
| #2 Earlier Production Quizzes | High | Low | 1st - Quick win |
| #1 Reverse Search by Meaning | High | Medium | 2nd - Solves immediate pain |
| #3 Sentence Generation Quiz | High | Medium-High | 3rd - Production practice |
| #4 Memory Palace Prompts | Medium | Medium-High | 4th - Experimental enhancement |

---

## Related Documentation
- Current Leitner implementation: `docs/DETERMINISTIC_LEITNER_IMPLEMENTATION.md`
- Quiz types: `api/internal/model/quiz.go`
- Leitner strategy: `api/internal/service/leitner_system_strategy.go`
