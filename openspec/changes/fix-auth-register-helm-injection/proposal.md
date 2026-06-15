## Why

The `auth-register` Helm Job directly interpolates Helm values (e.g., `.Values.authRegister.userName`, `.Values.appCode`, `authserverIngressHost`) into a `sh -c` bash here-doc, enabling command injection if an attacker can influence chart values (e.g., via GitOps/CI pipelines). This must be fixed to eliminate the code execution vector before it can be exploited.

## What Changes

- Remove the inline `sh -c` bash here-doc from `auth-register-job.yaml`
- Add a dedicated shell script `/data/bin/auth-register.sh` to the auth-register container image, mirroring the pattern already used by `apigw-register-job.yaml`
- Pass all Helm values as environment variables to the Job container instead of embedding them in the command string
- Add Helm value validation (`regexMatch`) in `values.yaml` / `_helpers.tpl` for user-controllable fields (`authRegister.userName`, `appCode`)

## Capabilities

### New Capabilities

- `auth-register-script`: A dedicated entrypoint script `/data/bin/auth-register.sh` that reads registration parameters from environment variables and executes the curl-based auth center registration safely, eliminating shell injection risk.

### Modified Capabilities

<!-- No existing spec-level capability requirements are changing. -->

## Impact

- **Files changed**:
  - `docs/support-file/helm/templates/authserver/auth-register-job.yaml` — replace inline command with `command: ["/data/bin/auth-register.sh"]` + `env:` block
  - Container image for auth-register — add `auth-register.sh` script to `/data/bin/`
  - `docs/support-file/helm/values.yaml` (optional) — add `regexMatch` validation for `authRegister.userName`
- **Behavior**: Identical registration logic; only the delivery mechanism changes (env vars instead of inline shell interpolation)
- **Breaking changes**: None — the Job behavior and API call are preserved exactly
