## Why

The `tpl_schema` sheet structure for GPU demand Excel import has been updated: each sheet now has both `fixed_headers` (fixed columns for sorting/aggregation) and `headers` (dynamic business columns), each with a `hidden` flag. The current header validation logic in `pkg/tools/excel/reader.go` uses `AllExcelHeaders()` which includes **all** headers with `field != "-"`, including hidden ones. This causes validation failures when hidden columns (e.g. formula-computed columns like "QPM峰值" or "预算卡数") do not have matching header text in the Excel file's header row. The validation must be updated to only check visible headers, and debug logging should be added to aid troubleshooting during integration.

## What Changes

- **Fix header validation**: Change `validateHeaders()` to use `AllVisibleHeaders()` instead of `AllExcelHeaders()`, so hidden columns in both `fixed_headers` and `headers` are excluded from Excel header row validation.
- **Fix JSON tag mismatch**: The API doc defines the field as `fixed_headers`, but `schema.go` uses `json:"fix_headers"`. Update the JSON tag to `json:"fixed_headers"` so the schema can be correctly unmarshaled from data-service responses.
- **Add debug logging**: Add `[jett]` prefixed debug logs in `validateHeaders()` that print, per sheet in one line: (1) all non-empty Excel header row values (with `%q` to reveal line breaks), (2) schema visible fixed_headers and headers (hidden=false, field!="-").
- **Update tests**: Update existing test cases in `reader_test.go` to cover the new behavior: hidden headers should be skipped during validation, not cause missing-column errors.

## Capabilities

### New Capabilities

_None — this change modifies existing validation logic only._

### Modified Capabilities

_No existing spec-level requirements are changing. This is a bug fix and implementation refinement within the existing `excel-import-gpu-demand` capability._

## Impact

- **Code**: `pkg/tools/excel/reader.go` (validateHeaders logic), `pkg/tools/excel/schema.go` (JSON tag fix), `pkg/tools/excel/reader_test.go` (test updates)
- **API**: No API contract change — the response schema already documents `fixed_headers` / `headers` with `hidden` field. This change makes the backend validation logic consistent with the documented schema.
- **Dependencies**: None
- **Risk**: Low — only affects server-side validation of uploaded Excel files. The JSON tag fix (`fix_headers` → `fixed_headers`) is **BREAKING** if any existing stored data uses the old key, but based on the API doc the correct key is `fixed_headers`.
