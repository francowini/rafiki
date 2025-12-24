# Telegram Moment Tracking - Use Cases

## UC-01: Start New Moment Session

**Actor**: Registered user with linked Telegram account

**Preconditions**:
- User has Telegram account linked (admin set `telegram_chat_id`)
- No active session exists

**Flow**:
1. User sends `/momento` to @RafikiBot
2. System creates new session (step_1, TTL 15 min)
3. System sends Step 1 prompt: "¿Qué pasó? Contanos qué estaba pasando..."
4. User responds with situation + thoughts
5. System calls Anthropic API to validate
6. If approved → advance to step 2
7. If needs_refinement → send feedback, stay on step 1

**Postconditions**:
- Session in step_1 or step_2
- `parsed_data.situacion` and `parsed_data.pensamientos` stored

---

## UC-02: Complete Full 6-Step Flow

**Actor**: User in active session

**Preconditions**:
- Active session exists
- User has completed previous steps

**Flow**:
1. User receives prompt for current step
2. User responds with text
3. System validates via Anthropic API
4. If approved:
   - Store parsed_data for step
   - If step < 6: advance to next step, send prompt
   - If step = 6: trigger UC-03 (Completion)
5. If needs_refinement:
   - Increment retry_count
   - If retry_count >= 2: auto-approve, advance
   - Else: send feedback, stay on step

**Steps**:
| Step | Field | Prompt |
|------|-------|--------|
| 1 | situacion, pensamientos | ¿Qué pasó? |
| 2 | sintomas_fisicos, emociones | ¿Qué sentiste? |
| 3 | conducta | ¿Qué hiciste? |
| 4 | consecuencias | ¿Qué pasó después? |
| 5 | tipo, descripcion (valores) | ¿Te acercó o alejó? |
| 6 | intensidad (0-10) | ¿Qué intensidad? |

**Postconditions**:
- All 6 steps completed
- Session marked for completion

---

## UC-03: Auto-Save Completed Moment

**Actor**: System (job worker)

**Preconditions**:
- Session in step_6
- Valid intensity (0-10) received

**Flow**:
1. Parse intensity from user response
2. Map session.parsed_data to Moment fields:
   - situation ← situacion
   - thoughts ← pensamientos
   - physicalSymptoms ← sintomas_fisicos
   - behavior ← conducta
   - consequences ← consecuencias
   - valuesReflection ← valores.descripcion
   - intensity ← intensidad
   - source ← "telegram"
3. Call `momentbus.Create(moment)`
4. Delete session from database
5. Send "✓ Momento guardado" to user

**Postconditions**:
- New moment created in database
- Session deleted
- User sees confirmation

---

## UC-04: Cancel Active Session

**Actor**: User in active session

**Preconditions**:
- Active session exists (any step)

**Flow**:
1. User sends `/cancel`
2. System deletes session (hard discard)
3. System sends "Momento descartado"

**Postconditions**:
- Session deleted
- No moment created
- No partial data saved

---

## UC-05: Session Timeout

**Actor**: System (cleanup job)

**Preconditions**:
- Session exists
- `last_activity` > 15 minutes ago

**Flow**:
1. Cleanup job runs every 5 minutes
2. Finds sessions with `last_activity + 15min < NOW()`
3. Deletes expired sessions
4. No notification sent to user (silent)

**User Re-entry**:
1. User sends message after timeout
2. System detects no active session
3. System sends "Tu sesión expiró. Usá /momento para empezar de nuevo."

**Postconditions**:
- Session deleted
- No moment created

---

## UC-06: Help Command

**Actor**: Any user with linked Telegram

**Preconditions**: None

**Flow**:
1. User sends `/ayuda`
2. System sends help message:
   ```
   Rafiki te ayuda a registrar momentos difíciles usando un análisis funcional.

   Comandos:
   • /momento - Empezar un nuevo registro
   • /cancel - Descartar el registro actual
   • /ejemplo - Ver un ejemplo de registro completo

   El registro tiene 6 pasos y dura unos 5 minutos.
   ```

**Postconditions**: None (no state change)

---

## UC-07: Example Command

**Actor**: Any user with linked Telegram

**Preconditions**: None

**Flow**:
1. User sends `/ejemplo`
2. System sends complete example conversation

**Postconditions**: None (no state change)

---

## UC-08: Unrecognized Message Outside Session

**Actor**: User without active session

**Preconditions**:
- No active session
- Message is not a command

**Flow**:
1. User sends random text (not /comando)
2. System checks for active session → none found
3. System sends "Usá /momento para empezar un nuevo registro."

**Postconditions**: None

---

## UC-09: Start While Session Active

**Actor**: User with active session

**Preconditions**:
- Active session exists

**Flow**:
1. User sends `/momento`
2. System detects active session
3. System sends "Ya tenés un momento en curso. Usá /cancel para descartarlo o seguí respondiendo."

**Postconditions**: Session unchanged

---

## UC-10: Unlinked User

**Actor**: Telegram user without linked account

**Preconditions**:
- User's `telegram_chat_id` not in database

**Flow**:
1. User sends any message to @RafikiBot
2. Webhook receives, looks up chat_id → not found
3. System sends "Tu cuenta de Telegram no está vinculada. Contactá al administrador."
4. Job not enqueued

**Postconditions**: None

---

## UC-11: Admin Links Telegram Account

**Actor**: Admin

**Preconditions**:
- User exists in database
- User's Telegram chat_id is known

**Flow**:
1. Admin calls POST /v1/admin/telegram/link:
   ```json
   {
     "user_id": "uuid",
     "telegram_chat_id": 123456789
   }
   ```
2. System updates user:
   - telegram_chat_id = 123456789
   - telegram_enabled = true
   - telegram_linked_at = NOW()
3. Returns success response

**Postconditions**:
- User can use @RafikiBot
- Linking is permanent until admin unlinks

---

## UC-12: Anthropic API Failure

**Actor**: System (job worker)

**Preconditions**:
- Active session
- Anthropic API unreachable or errors

**Flow**:
1. Job worker calls Anthropic API → error
2. Job retries (up to 3 attempts, exponential backoff)
3. If all retries fail:
   - Session remains in current step
   - Send "Hubo un problema técnico. Intentá de nuevo en un momento."
4. User's next message triggers retry

**Postconditions**:
- Session unchanged (same step)
- User can retry

---

## UC-13: Telegram Send Failure

**Actor**: System (job worker)

**Preconditions**:
- Anthropic validation complete
- Telegram API unreachable

**Flow**:
1. Job worker tries to send message → error
2. Job retries (up to 3 attempts)
3. If all retries fail:
   - Session state already updated
   - Log error for investigation
4. User's next message will receive current step prompt again

**Postconditions**:
- Session advanced (if validation passed)
- User may see duplicate prompts on retry

---

## UC-14: Invalid Intensity Response

**Actor**: User in step 6

**Preconditions**:
- Session in step_6
- User response not a valid 0-10 number

**Flow**:
1. User responds "mucho" or "11" or no number
2. Anthropic evaluates → needs_refinement
3. System sends "Necesito un número del 0 al 10. ¿Qué número le pondrías?"
4. User responds again
5. If valid → complete
6. If still invalid after 2 retries → assign 5 (midpoint), complete

**Postconditions**:
- Moment created with intensity (user-provided or default 5)

---

## UC-15: View Telegram-Created Moment in Web App

**Actor**: User viewing web dashboard

**Preconditions**:
- Moment created via Telegram (source = "telegram")

**Flow**:
1. User opens Rafiki web app
2. Navigates to Moments list
3. Sees all moments (web and Telegram) in same list
4. No visual distinction (MVP)
5. Can view, edit, delete same as web-created moments

**Postconditions**: None

---

## Error Handling Summary

| Error | Behavior |
|-------|----------|
| Unlinked user | "Tu cuenta no está vinculada" |
| Session timeout | Silent expire, re-entry message |
| Anthropic down | Retry 3x, then error message |
| Telegram down | Retry 3x, log error |
| Invalid intensity | Re-ask, then default to 5 |
| Concurrent session | Block /momento, suggest /cancel |
| Empty response | Re-ask with encouragement |
