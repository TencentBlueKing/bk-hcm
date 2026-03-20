## Context

The GPU demand Excel import endpoint (`POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import`) parses uploaded Excel files against a template schema (`tpl_schema`). The schema has evolved: each sheet now defines `fixed_headers` (fixed columns like year/month/GPU count) and `headers` (dynamic business columns), each with a `hidden` boolean flag.

Current state:
- `validateHeaders()` in `pkg/tools/excel/reader.go` calls `sheet.AllExcelHeaders()`, which returns **all** headers with `field != "-"` — including hidden ones.
- Hidden columns (e.g. "预算卡数" with `hidden: true`) may not have visible header text in the Excel file, causing false validation failures.
- The `Schema.Sheet.FixHeaders` JSON tag is `fix_headers` but the API response uses `fixed_headers`, causing an unmarshal mismatch.
- `AllVisibleHeaders()` already exists in `schema.go` but is **not used** for validation.

## Goals / Non-Goals

**Goals:**
- Header validation only checks visible headers (hidden != true, field != "-") from both `fixed_headers` and `headers`.
- JSON tag for `FixHeaders` matches the API contract (`fixed_headers`).
- Debug logging with `[jett]` prefix prints per-sheet: Excel header row values and schema visible headers, with line breaks rendered via `%q`.
- Tests updated to cover hidden-header-skip behavior.

**Non-Goals:**
- Changing the sheet-level validation logic (extra/missing sheets).
- Modifying the `ParseSheetRows` or `buildDetails` logic (those already handle hidden correctly).
- Changing `AllExcelHeaders()` behavior (still needed for data parsing).

## Decisions

### Decision 1: Use `AllVisibleHeaders()` for header validation

**Choice**: Replace `sheet.AllExcelHeaders()` with `sheet.AllVisibleHeaders()` in `validateHeaders()`.

**Rationale**: `AllVisibleHeaders()` already correctly filters for `field != "-" && !hidden`. The existing `AllExcelHeaders()` is still needed by `ParseSheetRows` and `buildDetails` for data extraction (hidden columns are read from Excel but excluded from `raw_data` later).

**Alternative considered**: Adding a `hidden` check inline in the validation loop — rejected because `AllVisibleHeaders()` already encapsulates the exact logic needed.

### Decision 2: Fix JSON tag from `fix_headers` to `fixed_headers`

**Choice**: Change `json:"fix_headers"` to `json:"fixed_headers"` in `schema.go`.

**Rationale**: The API documentation and data-service response both use `fixed_headers`. The current mismatch means `FixHeaders` would be empty after JSON unmarshal if the data source uses `fixed_headers`.

**Risk**: If any existing stored JSON uses `fix_headers`, this breaks deserialization. However, the API doc is authoritative and all template data goes through the documented schema.

### Decision 3: Debug logs with `[jett]` prefix

**Choice**: Keep existing debug logs in `validateHeaders()` with `[jett]` prefix. Use `%q` formatting to reveal any line breaks (`\n`) in Excel header values.

**Rationale**: Temporary debug logging aids integration testing. `%q` makes whitespace characters (newlines, tabs) visible without breaking log line structure.

## Risks / Trade-offs

- **[JSON tag change]** → Low risk. If stored template data uses `fix_headers` key, the `FixHeaders` slice will be empty. Mitigation: verify data-service response format before deploying.
- **[Debug logs in production]** → Debug `[jett]` logs should be removed before production release. Mitigation: treat them as temporary; remove in a follow-up cleanup.
- **[Visible-only validation may miss corrupt hidden columns]** → Acceptable trade-off. Hidden columns are formula-computed or system-managed; their correctness is not the user's responsibility.
