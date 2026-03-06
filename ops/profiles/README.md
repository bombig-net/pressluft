# Server Profiles

Profiles define the managed intent for server classes used by Pressluft.

A profile contract is declarative. It should describe:
- identity and version (`key`, `name`, `version`, `description`)
- support level and operator-facing reason (`support`)
- configure guarantee (`configure_guarantee`)
- image policy (`base_image`, `image_policy`)
- baseline convergence requirements (`baseline`)
- artifact references (`artifacts`)
- post-config verification checks (`verification`)

Contribution notes:
- Keep profile files provider-agnostic where possible.
- Treat profile version bumps as contract changes.
- Do not encode secrets in profiles.
- Validate changes against `ops/schemas/profile.schema.json`.

Current phase:
- Profiles are now the canonical source of operational intent.
- Runtime execution uses one explicit loading model: `dynamic-include-role`.
- `nginx-stack` is the only profile currently supported end-to-end.
- `openlitespeed-stack` and `woocommerce-optimized` remain as internal identifiers, but they are intentionally unavailable until their roles and verification contracts are real.
