# terraform-provider-readthedocs

**Repo:** https://github.com/the-robot-lives/terraform-provider-readthedocs

From-scratch Terraform/OpenTofu provider for the Read the Docs API v3.

## What

A Go provider (source address `the-robot-lives/readthedocs`, plugin protocol address `registry.terraform.io/the-robot-lives/readthedocs`) covering the full RTD API v3 surface. Not a fork of the abandoned 2022 `BarnabyShearer/readthedocs` provider or any MCP — client and schema are written from the official docs (https://docs.readthedocs.com/platform/stable/api/v3.html).

**Resources (write):**

| Resource | API |
|----------|-----|
| `readthedocs_project` | `POST/GET/PATCH /projects/` (no official DELETE) |
| `readthedocs_version` | `PATCH /projects/{p}/versions/{v}/` |
| `readthedocs_build` | `POST /projects/{p}/versions/{v}/builds/` (push/rebuild) |
| `readthedocs_sync_versions` | `POST /projects/{p}/sync-versions/` |
| `readthedocs_redirect` | `GET/POST/PUT/DELETE /projects/{p}/redirects/` |
| `readthedocs_environment_variable` | `POST/GET/DELETE /environmentvariables/` (no PATCH) |
| `readthedocs_subproject` | `POST/GET/DELETE /subprojects/` |
| `readthedocs_sharing` | Business `GET/POST/PATCH/DELETE /sharing/` |

**Data sources (read):** `project`, `projects`, `version`, `versions`, `build`, `builds`, `redirects`, `environment_variables`, `subprojects`, `translations`, `organization`, `organizations`, `organization_projects`, `organization_teams`, `remote_organizations`, `remote_repositories`, `embed`, `superproject`. List data sources expose `result_count` + `results_json` (raw API array).

Not implementable (not in public API v3): custom domains, incoming VCS webhooks, outgoing webhooks, documented project DELETE.

## Why

Noizu hosts docs for many portfolio projects on Read the Docs; managing projects, versions, redirects, environment variables, and subprojects in Terraform keeps that configuration versioned and reproducible instead of click-ops. No maintained provider existed, so this one was written against the v3 API directly.

## Getting Started

Prerequisites: Go (for build), Terraform or OpenTofu, a Read the Docs API token.

```hcl
provider "readthedocs" {
  # token    = var.rtd_token   # or READTHEDOCS_TOKEN env var
  # base_url = "https://app.readthedocs.com/api/v3"  # Business; default is .org
}
```

Auth uses `Authorization: Token …` — 60 req/min authenticated.

**House install is a local compile** into `~/.local/share/terraform/plugins` (same path as SigNoz, `noizu/foryou`, `noizu/google-marketing`); there is no HashiCorp Registry listing. GitHub `v*` releases carry checksums if a registry listing is ever wanted.

```bash
./scripts/build-provider.sh   # builds + installs to the local plugin dir (both registry.terraform.io and registry.opentofu.org layouts)
make compile && make test     # build and test via Makefile
```

`~/.terraformrc` wires it up — `dev_overrides` skips `init` download; `filesystem_mirror` is the versioned path OpenTofu uses for `noizu/*`:

```hcl
provider_installation {
  dev_overrides {
    "the-robot-lives/readthedocs" = "~/.local/share/terraform/plugins"
  }
  filesystem_mirror {
    path    = "~/.local/share/terraform/plugins"
    include = ["the-robot-lives/readthedocs"]
  }
  direct {}
}
```

Example config: [`examples/github-utils/main.tf`](examples/github-utils/main.tf) — needs `READTHEDOCS_TOKEN` (or `token` in the provider block); do not apply without a token.

## How It Works

- `internal/rtdapi/` — hand-written Read the Docs API v3 client.
- `internal/provider/` — Terraform plugin-framework resources and data sources.
- GPG signing key for releases: [`gpg-pubkey.asc`](gpg-pubkey.asc), fingerprint `2EABB783A4251C2A26FCD82E4CACEE95EB6E16D0`; `terraform-registry-manifest.json` is kept ready for a future registry publish.

## Docs

`docs/` carries PROJ-ARCH / PROJ-LAYOUT / PROJ-SCHEMA digests and full docs.

License: MPL-2.0 (see `LICENSE`).
