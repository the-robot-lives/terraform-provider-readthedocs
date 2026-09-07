# PROJ-LAYOUT.summary — terraform-provider-readthedocs

```text
terraform-provider-readthedocs/
├── main.go                    # Plugin entry point
├── internal/
│   ├── provider/              # Provider + 8 resources + 18 data sources
│   └── rtdapi/                # From-scratch RTD API v3 REST client (+ tests)
├── examples/github-utils/     # Example terraform config + local state
├── scripts/                   # build-provider.sh (local install)
├── .github/workflows/         # ci.yml, release.yml
├── .goreleaser.yml            # Release config
├── gpg-pubkey.asc             # Release signing key
├── terraform-registry-manifest.json
├── go.mod / go.sum
├── Makefile
├── CLAUDE.md
├── README.md / merge-notes.md / TODO.md / LICENSE
└── docs/                      # PROJ-LAYOUT.md, PROJ-SCHEMA.md, PROJ-ARCH.md
```
