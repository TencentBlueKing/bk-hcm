## 1. Fix JSON Tag Mismatch

- [x] 1.1 In `pkg/tools/excel/schema.go`, change `FixHeaders` JSON tag from `json:"fix_headers"` to `json:"fixed_headers"` to match the API contract

## 2. Fix Header Validation Logic

- [x] 2.1 In `pkg/tools/excel/reader.go` `validateHeaders()`, replace `sheet.AllExcelHeaders()` with `sheet.AllVisibleHeaders()` so hidden headers are excluded from validation
- [x] 2.2 Update the debug log block for `fix_headers` to also filter by `!h.Hidden` (only log visible fixed_headers), keeping `%q` formatting for line break visibility

## 3. Update Tests

- [x] 3.1 In `pkg/tools/excel/reader_test.go`, rename test case `hidden_headers_still_validated` to `hidden_headers_skipped_in_validation` and update its expectation: hidden header NOT in Excel should still pass validation
- [x] 3.2 Add new test case `hidden_fixed_header_not_in_excel_passes`: schema has a hidden fixed_header with a real field, Excel does NOT have that header text — validation should pass
- [x] 3.3 Add new test case `missing_visible_fixed_header_fails`: schema has a visible fixed_header, Excel is missing that header — validation should fail with "missing columns"
- [x] 3.4 Add new test case `mixed_visible_hidden_headers`: schema has mix of visible and hidden headers in both fixed_headers and headers — validation checks only visible ones

## 4. Verify & Cleanup

- [x] 4.1 Run `go test ./pkg/tools/excel/...` and verify all tests pass
- [x] 4.2 Run `go vet ./pkg/tools/excel/...` to check for static analysis issues
