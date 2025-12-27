# Product TODO

## Objectives Module

### High Priority

- [ ] **Implement Activity Endpoint** - The frontend calls `GET /v1/objetivos/{objetivo_id}/activity?year={year}` but this endpoint doesn't exist in the backend. The `react-activity-calendar` heatmap component requires:
  - All 365 days of the year with `{ date: "YYYY-MM-DD", count: number, level: 0-4 }`
  - `totalCompletions`: Total completed records
  - `streakDays`: Current streak
  - `longestStreak`: Best streak achieved

  Implementation needed:
  - Add route in `objetivoapp/route.go`
  - Create handler that aggregates records by date
  - Calculate streak statistics


- [ ] **Fix bad login message** - query: email[{ test@rafiki.lat}]: query: email[{ test@rafiki.lat}]: db: user not found


### UX/Design Questions

- [ ] **Completed Objectives Display** - We don't know how to visually handle objectives with `status: completado` to make them look nice. Questions to resolve:
  - Should completed objectives be grayed out or have a special visual treatment?
  - Should they show a celebration/success state?
  - Should they be moved to a separate "Completed" section?
  - Should the heatmap still be interactive or read-only?
  - How to differentiate from archived objectives?

---

## Telegram Moment Tracking

### High Priority (Deliverable 5c - Next)

#### 1. Moment Creation Integration

**Status**: Next Deliverable
**Scope**: `app/jobs/telegrammessage/completion.go`

- [ ] Map session `parsed_data` to `momentbus.NewMoment`
- [ ] Add `emotions` field to Moment model (requires migration)
- [ ] Add `source` field to Moment model (`"web"` | `"telegram"`)
- [ ] Call `momentbus.Create()` after Step 6 completion
- [ ] Delete session after successful moment creation
- [ ] Handle partial failures (moment created but session delete fails)

**Field Mapping**:
| Session Field | Moment Field |
|---------------|--------------|
| `step_1.situacion` | `Situation` |
| `step_1.pensamientos` | `Thoughts` |
| `step_2.sintomas_fisicos` | `PhysicalSymptoms` |
| `step_2.emociones` | `Emotions` (NEW) |
| `step_3.conducta` | `Behavior` |
| `step_4.consecuencias` | `Consequences` |
| `step_5.descripcion` | `ValuesReflection` |
| `step_6.intensidad` | `Intensity` |

---

### Medium Priority (Phase 2)

#### 2. Malformed JSON Fallback

**Status**: Deferred
**Reason**: Current implementation treats malformed JSON as critical error (fails job)

**Future Enhancement**:
- [ ] Implement regex-based extraction as fallback
- [ ] Try to extract status/feedback from plain text response
- [ ] Log malformed responses for prompt tuning
- [ ] Consider structured output API when available in Anthropic

**Current Behavior**: Job fails, River retries 3 times, then user sees error message

---

#### 3. Moment Date Parsing

**Status**: Deferred
**Reason**: Complex NLP required, using `time.Now()` for MVP

**Current Implementation**:
```go
MomentDate: time.Now() // When session completed
```

**Future Enhancement**:
- [ ] Add optional "When did this happen?" question
- [ ] Parse relative dates: "yesterday", "this morning", "2 hours ago"
- [ ] Parse absolute dates: "on Monday", "December 15"
- [ ] Consider using AI to extract datetime from user text

**Rationale**: Most users track moments immediately or within minutes. Timestamp reflects "when I registered" not "when it happened".

---

#### 4. Crisis Detection & Safety Protocol

**Status**: Deferred
**Reason**: Requires professional review, legal considerations

**Why This Matters**:
This is a **mental health tool**. Users may express:
- Self-harm ideation
- Suicidal thoughts
- Severe distress

**Future Implementation**:
- [ ] Add crisis keyword detection to AI prompts
- [ ] Define emergency response flow:
  1. Send empathetic message with crisis resources
  2. Flag moment with `crisis_flagged: true`
  3. Notify admin via email
  4. Continue session (don't abandon user)
- [ ] Partner with licensed mental health professional
- [ ] Legal review for liability

**Crisis Resources (Argentina)**:
- Centro de Atencion al Suicida: 135 (free, 24/7)
- Help line: [to be determined]

**Temporary Mitigation**:
- [ ] Add safety disclaimer to `/ayuda` command response

---

#### 5. AI Insights Storage

**Status**: Deferred
**Reason**: Not needed for MVP, requires additional database column

**Future Enhancement**:
- [ ] Add `ai_insights JSONB` column to moments table
- [ ] Store Claude's raw response for each step
- [ ] Enable future pattern analysis
- [ ] Power weekly AI summaries

**Schema Change**:
```sql
ALTER TABLE moments ADD COLUMN ai_insights JSONB DEFAULT '{}'::jsonb;
```

---

### Low Priority (Phase 3+)

#### 6. Positive Moment Tracking

**Status**: Future Phase
**Reference**: Overview doc mentions "Positive moments in Phase 2"

- [ ] Add neutral prompts for positive moments
- [ ] Different session type: `positive_moment_tracking`
- [ ] Consider different step flow (approach-focused)

---

#### 7. WhatsApp Integration

**Status**: Future Phase
**Reference**: "Same architecture" per overview doc

- [ ] Evaluate WhatsApp Business API
- [ ] Reuse session domain (add `whatsapp` session type)
- [ ] Create WhatsApp webhook handler
- [ ] Adapt message formatting for WhatsApp

---

#### 8. Weekly AI Summaries

**Status**: Future Phase

- [ ] Scheduled job to analyze weekly moments
- [ ] Identify patterns (time of day, triggers, behaviors)
- [ ] Generate personalized insights
- [ ] Send summary via Telegram

---

#### 9. Frontend Moment Source Filter

**Status**: Future Phase
**Prerequisite**: Deliverable 5c (source field added)

- [ ] Add `source` filter to moment list
- [ ] Visual indicator for Telegram-created moments
- [ ] Display AI insights in moment detail view

---

### Technical Debt

#### 10. Prompt Validation at Startup

**Status**: Should be added to tests

- [ ] Test that all steps 1-6 are present in `prompts/steps.yaml`
- [ ] Validate required fields (question, instruction, parse_fields)
- [ ] Fail fast if prompts malformed

```go
func TestLoadPrompts_AllStepsPresent(t *testing.T) {
    prompts := LoadPrompts()
    for i := 1; i <= 6; i++ {
        if _, ok := prompts.Steps[i]; !ok {
            t.Errorf("missing step %d in prompts.yaml", i)
        }
    }
}
```

---

#### 11. AI Response Field Validation

**Status**: Should be added

- [ ] Validate `parsed_data` contains expected fields from `parse_fields`
- [ ] Log warning if AI returns unexpected fields
- [ ] Consider strict mode for production

---

#### 12. Centralized Logging (Future)

**Status**: Deferred for 10-user testing phase
**Current**: Logs only via SSH

When user volume increases:
- [ ] Set up BetterStack Logs (free tier)
- [ ] Configure syslog driver in docker-compose
- [ ] Create dashboards for key metrics
- [ ] Set up alerts for error rates

---

## Completed Items

### Telegram Moment Tracking
- [x] Deliverable 1: Anthropic Client (`foundation/anthropic/`)
- [x] Deliverable 2: Database Schema (`telegram_sessions` table)
- [x] Deliverable 3: Session Domain (`telegramsessionbus`)
- [x] Deliverable 4: Webhook Handler (`telegramapp`)
- [x] Deliverable 5a: Job Worker Core Structure
- [x] Deliverable 5b: Anthropic Integration

---

## Notes

### Per-Step Retry Tracking

**Decision**: Track retries per step (not global)
**Implementation**: `RetryCount` resets to 0 when advancing to next step
**Already Supported**: `telegramsessionbus.AdvanceStepWithData()` resets retry count

### Intensity Default

**Decision**: Default to 5 (not null) with educational message
**Rationale**: Self-awareness of emotional intensity is a core ACT skill
**Message**: "Assigned intensity 5 (medium). With practice, it will be easier to identify these differences."

### Compassionate Auto-Approval

**Decision**: Auto-approve after 2 retries with validating message
**Rationale**: Prevents frustration, teaches self-compassion
**Message**: "I see this step is difficult right now, and that's okay. Let's continue with what you could identify."
