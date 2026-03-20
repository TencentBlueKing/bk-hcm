## ADDED Requirements

### Requirement: Validate only visible headers against Excel header row

The system SHALL validate Excel file headers only against **visible** columns from both `fixed_headers` and `headers` in the template schema. A visible column is defined as one where `field != "-"` AND `hidden != true`. Hidden columns (e.g. formula-computed fields like "预算卡数", "QPM峰值") SHALL be excluded from header row validation.

#### Scenario: Hidden fixed_header is skipped during validation

- **GIVEN** a schema sheet with `fixed_headers` containing a header with `hidden: true` and `field: "C"`
- **WHEN** the Excel file does not contain this header's name in the header row
- **THEN** validation SHALL pass without reporting a missing column error

#### Scenario: Visible fixed_headers are validated before regular headers

- **GIVEN** a schema sheet with visible `fixed_headers` (hidden=false) at columns A, B and `headers` at columns D, E
- **WHEN** the Excel header row contains values matching all visible fixed_headers and all visible headers
- **THEN** validation SHALL pass

#### Scenario: Missing visible fixed_header causes validation failure

- **GIVEN** a schema sheet with a visible `fixed_header` named "年份" at column A
- **WHEN** the Excel header row does not contain "年份"
- **THEN** validation SHALL fail with an error reporting the missing column

### Requirement: JSON tag alignment for fixed_headers

The `FixHeaders` field in the `Sheet` struct SHALL use JSON tag `fixed_headers` to match the API contract defined in the `tpl_schema` response format.

#### Scenario: Schema unmarshals correctly with fixed_headers key

- **GIVEN** a JSON payload with `"fixed_headers": [...]` in a sheet definition
- **WHEN** the JSON is unmarshaled into the `Sheet` struct
- **THEN** the `FixHeaders` field SHALL contain the deserialized header list

### Requirement: Debug logging for header validation

The `validateHeaders` function SHALL log per-sheet debug information with `[jett]` prefix, including:
1. All non-empty Excel header row values with `%q` formatting to reveal line breaks
2. All schema visible headers (fixed_headers + headers where hidden=false and field!="-")

#### Scenario: Excel header with embedded newline is logged with escape sequence

- **GIVEN** an Excel header cell value containing a newline character (e.g. "QPM峰值\n(需同时填写)")
- **WHEN** header validation runs and logs the Excel headers
- **THEN** the log output SHALL display the header as a quoted string with `\n` visible (e.g. `"QPM峰值\n(需同时填写)"`)
