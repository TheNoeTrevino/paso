# Paso - Remaining Implementation Tasks

## Project Status

**Phases 1-7: Complete ✅**

See `HANDOFF.md` for complete implementation details of what's been built.

## What's Left

### Phase 9: Data Persistence & Reloading Tests

**Goal**: Verify that all data operations persist correctly and reload properly.

**Tasks:**
1. Write integration tests for persistence (`internal/database/persistence_test.go`)
2. Test scenarios:
   - Task CRUD operations persist across app restart
   - Column CRUD with linked list integrity
   - Task movement between columns
   - Column insertion (after current) maintains order
   - Cascade deletion (column + all tasks)
   - Transaction rollback on errors
   - Migration idempotency
   - Empty database reload (default columns)

**Definition of Done:**
- All operations immediately save to database
- Restarting app loads exact previous state
- Linked list integrity maintained
- Timestamps update correctly
- No data loss on unexpected exit

---

### Phase 10: Polish & User Experience (Optional)

**Goal**: Add final touches for production-ready feel.

**Enhancements to Consider:**

1. **Task Counts in Headers**
   ```go
   // components.go
   header := fmt.Sprintf("%s (%d)", column.Name, len(tasks))
   ```

2. **Relative Timestamps**
   ```go
   // Show "2h ago" for recently updated tasks
   if timeSince < 24*time.Hour {
       timeStr = fmt.Sprintf("%dh ago", int(timeSince.Hours()))
   }
   ```

3. **Help Screen** (`?` key)
   ```go
   // Show all keyboard shortcuts in centered dialog
   case "?":
       m.mode = HelpMode
   ```

4. **Status Bar**
   ```go
   // Bottom bar: "3 columns | 12 tasks | Press ? for help"
   ```

5. **Better Error Messages**
   ```go
   // Show error banner at top when operations fail
   if m.errorTimeout > 0 {
       errorBanner := errorStyle.Render("⚠ " + m.errorMessage)
   }
   ```

6. **Empty State Messages**
   ```go
   // When column has no tasks
   content += emptyStyle.Render("No tasks")
   ```

**Definition of Done:**
- UI feels polished and professional
- Help is easily accessible
- Empty states are clear
- Errors are user-friendly

---

## Current State

**Working Features:**
- Database with linked-list columns
- Full CRUD for tasks (a/e/d)
- Full CRUD for columns (C/R/X)
- Navigation (hjkl, arrows)
- Viewport scrolling ([ ])
- Task movement (< >)

**Test Coverage:**
- 9/9 repository tests passing (linked list operations)
- Need: Integration tests for persistence

**Next Steps:**
1. Implement Phase 9 tests
2. Optionally add Phase 10 polish
3. Done! 🎉
