# THREAT-MODEL.summary — terraform-provider-readthedocs

**Local CLI plugin, not a service** — no ingress/listeners/persistence. Attack surface: env/config
token input, outbound HTTPS to Read the Docs API v3, local tfstate + plugin mirror, GPG-signed
release pipeline.

- **Crown jewels**: RTD API token (full account control); tfstate carrying env-var values + sharing
  passwords in plaintext.
- **Trust boundaries**: user machine ↔ RTD API (TLS); provider ↔ local filesystem; CI ↔ release
  consumers (SHA256SUMS + GPG).
- **Register (8)**: T-001 High tfstate-in-git (mitigated, gitignored) · T-002 Medium plaintext
  secrets in tfstate (partial — `Sensitive` flags only) · T-003 Medium token via env (accepted) ·
  T-004 Medium CI actions tag-pinned not SHA-pinned (**open**) · T-005 Low error bodies echo API
  response ≤500B (accepted) · T-006 Low no cert pinning (accepted) · T-007 Low no retry/backoff on
  429s (accepted) · T-008 Low no provider-side audit log (accepted).
- **Coverage**: 1 mitigated · 1 partial · 1 open (T-004: SHA-pin the four workflow actions) ·
  5 accepted.
- **Residual risk**: T-002 and T-004 are the plausible rides; both bounded (throwaway example
  project; signing key only in encrypted GH secrets).
