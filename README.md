# dkron-with-auth

**Dkron – Distributed, fault-tolerant job scheduling system with optional Web UI authentication.**

This repository is a community fork of Dkron that adds a **lightweight, optional authentication layer for the Web UI**, while keeping the **server, CLI, agents, and API fully compatible with upstream behavior by default**.

---

## Overview

This fork introduces **UI-only authentication** that can be enabled via configuration without impacting existing job execution or automation workflows.

**Key properties:**
- Authentication applies **only to the Web UI**
- Disabled by default (matches upstream OSS behavior)
- No API, CLI, or agent changes unless explicitly enabled

---

## Highlights

- UI-only authentication using **HMAC-signed session cookies**
- Configurable and **opt-in**
- Session TTL support
- No breaking changes to existing flows

---

## UI Auth Modes

- **Disabled (default)**  
  `ui-auth-enabled: false`  
  → UI remains open, identical to upstream behavior

- **Session authentication**  
```

ui-auth-enabled: true
ui-session-enabled: true

```
→ UI requires login; a signed `dkron_ui_session` cookie is issued with a TTL

---

## Default Credentials ⚠️

When UI authentication is enabled, the **default credentials are**:

- **Username:** `admin`  
- **Password:** `secret`

> ⚠️ These defaults are provided for **local development only** and **must be changed before any non-local or shared deployment**.

---

## Configuration

Credentials and session settings are configured in:

```

config/auth.yaml

````

Example:

```yaml
ui-auth-enabled: true
ui-auth-username: admin
ui-auth-password: change-me
ui-session-enabled: true
ui-session-ttl: 15m
ui-session-secret: replace-this-secret
````

### Notes

* Setting `ui-auth-enabled: false` restores original UI behavior
* `ui-session-ttl` uses Go duration format (e.g. `10m`, `1h`)
* Session cookie is `HttpOnly` and HMAC-signed using `ui-session-secret`

---

## Running

Minimal example (reads `auth.yaml`):

```bash
go run . agent --server --bootstrap-expect 1
```

Access the UI:

* `http://localhost:8080/ui/`

When session auth is enabled:

* A login page is shown
* Successful login sets a session cookie
* User is redirected back to the originally requested UI route

---

## Security Notes

* UI session cookies are HMAC-signed and time-bound
* For HTTPS deployments, consider enabling:

  * `Secure`
  * `SameSite` attributes
* This feature **only protects UI routes**
* API authorization is unchanged; use ACLs or enterprise features if required

---

## Compatibility

* API, CLI, and agent behavior remain unchanged
* Existing deep links continue to work
* Invalid or missing UI session redirects to login, then returns to the original route

---

## License

* Licensed under the **GNU Lesser General Public License v3.0 (LGPL-3.0)**
* This repository contains modifications to an LGPL-licensed project
* See `LICENSE` for full terms

---

## Attribution

This project is a fork of the original Dkron project.
It is **not officially affiliated** with the upstream maintainers.


