# dkron-with-auth

This fork of Dkron adds an optional, lightweight authentication layer for the Web UI while keeping the server, CLI and API compatible with the original flows by default.

Highlights:
- UI-only auth with HMAC-signed, TTL session cookie
- Toggleable via configuration; defaults to original OSS behavior
- No changes to API routes, CLI, or agents unless you enable UI auth

## UI Auth Modes
- Disabled (default): `ui-auth-enabled: false` → UI is open as in upstream OSS
- Session auth: `ui-auth-enabled: true` and `ui-session-enabled: true` → UI is protected; login sets `dkron_ui_session` cookie with TTL

## Configuration
Place UI auth settings in `config/auth.yaml` (or `dkron/config/auth.yaml`). Only defined keys are applied.

Example:
```
ui-auth-enabled: true
ui-auth-username: admin
ui-auth-password: secret
ui-session-enabled: true
ui-session-ttl: 15m
ui-session-secret: devsecret
```

Notes:
- `ui-auth-enabled: false` restores original UI behavior (no auth)
- `ui-session-ttl` uses Go duration format (e.g., `10m`, `1h`)
- Session cookie is `HttpOnly` and signed with `ui-session-secret`

## Run
Minimal (reads `auth.yaml`):
```
go run . agent --server --bootstrap-expect 1
```

Access the UI:
- `http://localhost:8080/ui/`
- When session auth is enabled, a login page is shown; successful login sets the cookie and reloads the requested route

## Security
- UI session cookie is HMAC-signed and expires according to `ui-session-ttl`
- Consider adding `Secure` and `SameSite` attributes if serving over HTTPS
- This auth layer protects UI pages only; use ACLs/Dkron Pro for API authorization

## Compatibility
- API, CLI and agent behavior unchanged unless you set `ui-auth-enabled: true`
- Deep-links to UI routes work; missing/invalid cookie shows the login page and returns to the original route after login

## Acknowledgements
- Based on the upstream Dkron project by Distribworks

## License
- Licensed under the GNU Lesser General Public License v3.0 (LGPL-3.0)
- See `LICENSE` for the full text
- Starts Mailpit (simulating the CI service container)
- Runs tests with the same configuration as GitHub Actions
- Provides clear pass/fail results
- Allows you to inspect emails in the Mailpit UI

See [.github/TESTING.md](.github/TESTING.md) for more information about CI testing.

### Frontend development

Dkron dashboard is built using [React Admin](https://marmelab.com/react-admin/) as a single page application.

To start developing the dashboard enter the `ui` directory and run `npm install` to get the frontend dependencies and
then start the local server with `npm start` it should start a new local web server and open a new browser window
serving de web ui.

Make your changes to the code, then run `make ui` to generate assets files. This is a method of embedding resources in
Go applications.

### Resources

Chef cookbook
https://supermarket.chef.io/cookbooks/dkron

Python Client Library
https://github.com/oldmantaiter/pydkron

Ruby client
https://github.com/jobandtalent/dkron-rb

PHP client
https://github.com/gromo/dkron-php-adapter

Terraform provider
https://github.com/bozerkins/terraform-provider-dkron

Manage and run jobs in Dkron from your django project
https://github.com/surface-security/django-dkron

## Contributors

<a href="https://github.com/distribworks/dkron/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=distribworks/dkron" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## Get in touch

- Twitter: [@distribworks](https://twitter.com/distribworks)
- Chat: https://gitter.im/distribworks/dkron
- Email: victor at distrib.works
