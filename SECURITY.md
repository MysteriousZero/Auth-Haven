# Security policy

## Project status

Auth Haven is under active development and has no published supported-release matrix. Security fixes are currently applied to the default branch. Do not assume the repository is production-ready without an independent security review.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability or include exploit details in public discussions.

Use GitHub's private vulnerability reporting feature for `MysteriousZero/auth-haven` if it is enabled. If it is unavailable, contact the repository owner through a private channel shown on the owner's GitHub profile and provide only enough public context to establish a secure communication channel.

Include, when possible:

- A concise description and affected component
- Reproduction steps or a minimal proof of concept
- Security impact and required preconditions
- Affected commit or version
- Suggested mitigation, if known

Do not access data you do not own, degrade services, persist access, or disclose secrets while investigating.

## Response expectations

There is currently no guaranteed response or remediation timeline. The maintainer should acknowledge valid reports privately, coordinate remediation and disclosure, and credit reporters when requested and appropriate.

## Security expectations for contributors

- Never commit secrets or real identity data.
- Treat `.env`, signing keys, MFA keys, database credentials, reset tokens, invitation tokens, and refresh tokens as sensitive.
- Preserve anti-enumeration behavior in authentication and password recovery.
- Do not suppress CodeQL, dependency-review, or secret-scanning findings without the evidence, owner, and review date required by the [CI policy](docs/ci.md#security-triage-and-suppression).
- Enforce tenant scope in services and repositories; never trust a path parameter alone for authorization.
- Use cryptographically secure randomness and established libraries rather than custom cryptography.
- Store opaque credentials as hashes and encrypt recoverable MFA secrets at rest.
- Avoid logging passwords, raw tokens, MFA codes, or decrypted secrets.
- Add regression tests for security fixes without publishing weaponized exploit details.

See [docs/configuration.md](docs/configuration.md) and [docs/security.md](docs/security.md) for configuration and security architecture.
