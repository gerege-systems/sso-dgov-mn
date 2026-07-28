# eID Login

## The Relying Party model

The platform is a Relying Party (RP) of the eID Mongolia identity provider
(`https://eidmongolia.mn/v3`). The RP client is `backend/pkg/eid/eid.go`. It
speaks a Smart-ID-compatible v3 API (ACSP_V2), authenticating with
`Authorization: Bearer <rp_sk_…>` plus `relyingPartyUUID` / `relyingPartyName`
in the body.

The requested certificate level floor is `ADVANCED` — deliberately the minimum,
since raising it to `QUALIFIED` would lock out citizens whose auth certificate is
only ADVANCED.

Wire protocol:

| Call | Purpose |
|------|---------|
| `POST {base}/authentication/device-link/anonymous` | QR / deep-link login |
| `POST {base}/authentication/notification/etsi/PNOMN-{civil}` | national-ID push login |
| `GET {base}/session/{sessionID}?timeoutMs=25000` | long-poll state |

## Three ways to start

| Endpoint | Usecase | Mode |
|----------|---------|------|
| `POST /api/v1/auth/eid/start` | `EIDStart` | QR / mobile deep-link |
| `POST /api/v1/auth/eid/start-id` | `EIDStartByNationalID` | national-ID push |
| `POST /api/v1/auth/eid/poll` | `EIDPoll` | long-poll to completion |

**QR vs deep-link.** `EIDStart` calls `eid.QRInitiate`. The request may carry an
optional `callbackUrl`:

- empty → **cross-device** (desktop QR: the browser polls);
- present (`<origin>/auth/eid/callback`) → **same-device** (mobile App2App: the
  browser is returned to that URL after the phone approves).

The response returns `session_id`, `device_link_url`, `verification_code`,
`expires_at`.

**National-ID push.** `EIDStartByNationalID` calls `eid.Initiate(nationalID, …)`
— eID pushes an approval prompt straight to the device(s) registered to that
national ID, so no QR is needed. The national ID is never logged (only a
`has_national_id` boolean). A per-request 16-byte crypto-random **nonce** guards
against IdP replay.

## The long-poll session lifecycle

The client posts `session_id` to `/auth/eid/poll`. The usecase calls
`uc.eid.Session(ctx, sessionID, 25000)` — a **25 s** hold, kept below the eID
client's 30 s HTTP timeout. The four platform states are `RUNNING`, `COMPLETE`,
`EXPIRED`, `REFUSED`.

On `COMPLETE`:

1. **Derive the subject.** A public RP does not receive `national_id`, so
   `civil_id` is the durable key (falling back to `national_id` for privileged
   RPs). Both empty → reject.
2. **Build the user** via `domain.NewEIDUser(...)` and attach PKI details read
   from the login certificate: `DocumentNumber`, `CertSerial`, `CertNotBefore`,
   `CertNotAfter`, `CertIssuer`, `CertKeyType`.
3. **Upsert** via `users.UpsertFromEID` — the Postgres upsert keys on a partial
   unique index `ON CONFLICT (lower(civil_id)) WHERE civil_id IS NOT NULL`; the
   username is `eid_<civil_id>`.
4. **Link a pending Google account** if a `google_link_token` was passed
   (non-fatal).
5. **MFA gate** — if the user is a super admin, no session is minted; a
   super-admin MFA challenge is started and `{ MFARequired: true, MFAToken }` is
   returned. See [Super Admin MFA](super-admin-mfa.md).
6. Otherwise mint an access + refresh pair and return
   `{ User, AccessToken, RefreshToken }`.

```mermaid
sequenceDiagram
    participant B as Browser (BFF)
    participant API as DAN backend
    participant IdP as eID Mongolia
    participant R as Redis
    participant DB as Postgres

    B->>API: POST /auth/eid/start {callbackUrl?}
    API->>IdP: QRInitiate(displayText, callbackUrl, nonce)
    IdP-->>API: session_id, device_link_url, verification_code
    API-->>B: show QR / deep-link

    loop every ~2.5s until terminal
        B->>API: POST /auth/eid/poll {session_id}
        API->>IdP: Session(session_id, timeoutMs=25000)
        IdP-->>API: RUNNING | COMPLETE | EXPIRED | REFUSED
    end

    Note over API,IdP: On COMPLETE
    API->>API: subject = civil_id; read cert/PKI details
    API->>DB: UpsertFromEID (ON CONFLICT lower(civil_id))
    alt super admin
        API->>R: SET superadmin_mfa:<token> (5m)
        API-->>B: {COMPLETE, MFARequired:true, MFAToken}
    else regular user
        API->>API: GenerateTokenPair(...)
        API->>R: SET refresh:<jti>
        API-->>B: {COMPLETE, User, AccessToken, RefreshToken}
    end
```

!!! note "Trusted-transport model"
    The IdP is a TLS-protected authoritative source, so the platform currently
    trusts the `COMPLETE` response. Verifying the ACSP_V2 signature against the
    certificate is flagged in the code as a future optional hardening.

## Identity / PKI data read back

Beyond login, the eID usecase exposes rich PKI data for the logged-in citizen
under `/api/v1/users/me/eid` (all resolve the citizen's ETSI id as
`PNOMN-<CIVIL_ID>`):

| Endpoint | Data |
|----------|------|
| `GET …/eid/summary` | eID account summary |
| `GET …/eid/certificates` | certificates |
| `GET …/eid/devices` | registered devices |
| `GET …/eid/activity` | RP-scoped auth/sign history |
| `GET/POST/DELETE …/eid/organizations` | linked organizations the citizen represents |
| `…/eid/organizations/{regNo}/signers` | list / add / resend / remove org signers |

Registering an organization performs a XYP lookup of CEO/founder/stakeholder
reg-numbers. Added signers get an eID sign-push and stay `PENDING` until they
approve with their PIN. A missing `PKI_READ` grant maps to `Forbidden`.

## Google account-linking

Google is **not** a standalone login; it links to (or logs in as) an
eID-verified person. Usecase: `auth_google.go`.

- **`POST /auth/google`** exchanges the OAuth `code`, then looks up
  `GetByGoogleSub`:
    - **already linked** → refreshes the stored Google profile and (unless the
      user is a super admin) mints a session, returning `{ Linked: true, Login }`.
    - **first time** → creates a short-lived `LinkToken` (15 min TTL), stores the
      full Google profile in Redis under `google_link:<token>`, and returns
      `{ Linked: false, LinkToken, Email }`. The client then runs an eID login,
      passing this token as `google_link_token`.
- **Linking on eID completion** (`linkGoogleIfPending`) is entirely non-fatal —
  the eID login always succeeds even if linking fails (e.g. that Google account
  is already bound to someone else).
- **`DELETE /auth/google/link`** unlinks; it is the one Google endpoint behind
  `authMiddleware`. Re-linking is only possible through the login flow.
