# API Explorer

The interactive OpenAPI explorer below is rendered from the backend's generated
spec (`backend/docs/swagger.json`). Use the [REST API](index.md) reference for
the authoritative, code-verified endpoint list — the generated spec has known
gaps and stale entries (flagged there).

!!! warning "Spec caveats"
    - Some groups (RBAC, Provider, admin-users) are missing from the spec.
    - Some password/OTP/registration auth paths in the spec were **removed** from
      the code (eID is the sole login method).
    - The spec `host` is the dev default `localhost:8080`.

!!swagger ../assets/swagger.json!!
