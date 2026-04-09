# Security Specification

This document details the security mechanisms, cryptographic algorithms, and defensive measures implemented in Auth Haven.

## 1. Cryptography & Hashing

We enforce strong, industry-standard cryptographic algorithms for all sensitive data stored in the database.

*   **Passwords:** Hashed using `argon2id`. This provides superior resistance against GPU cracking and side-channel attacks compared to standard bcrypt.
    *   *Parameters:* `time=1`, `memory=64MB`, `threads=4`, `keyLen=32`, `saltLen=16`.
*   **Refresh, Reset & Invitation Tokens:** Generated as high-entropy secure random strings (minimum 32 bytes, hex/base64 encoded). The raw tokens are sent to the user on creation, but **only their SHA-256 hashes are stored in the database**. This guarantees that even a full database leak does not expose active persistent sessions.
*   **TOTP Secrets & MFA Phone Numbers:** Encrypted at rest in the database using AES-256-GCM. The encryption key is injected via environment variables. This protects PII from database leaks.

## 2. Token Standards (Access vs. Refresh)

*   **Access Tokens:** Short-lived JWTs (JSON Web Tokens). 
    *   *Lifespan:* 15 minutes.
    *   *Signing Algorithm:* EdDSA (Ed25519) for asymmetric signing. The API holds the private key, while other microservices/clients can use the public key to quickly verify token signatures locally.
    *   *Claims:* `sub` (UserID), `iss`, `aud`, `exp`, `tenant_id`, `role_id`.
*   **Refresh Tokens:** Long-lived opaque tokens.
    *   *Lifespan:* 7 days (or 30 days if "remember me" is used).
    *   *Operation:* Must be validated against the database table `refresh_tokens`. Every refresh action triggers token rotation (a new Access+Refresh token pair is issued and the old Refresh token is revoked) to prevent token replay attacks.

## 3. Defensive Measures

### Rate Limiting & Brute Force Protection

To mitigate enumeration and brute-force attacks, the gateway/interceptor layers enforce strict rate limiting using Redis sliding windows:

*   **`/v1/auth/login`**: Max 5 attempts per 5 minutes per IP. Max 10 attempts per 5 minutes per Email.
*   **`/v1/auth/password/forgot`**: Max 3 attempts per 15 minutes per IP.
*   **`/v1/auth/mfa/verify`**: Max 5 attempts per 15 minutes per IP.

### Account Lockouts

We do NOT permanently lock accounts automatically to prevent Denial of Service (DoS) attacks on legitimate users. Instead, successive failed login attempts trigger exponential backoff / CAPTCHA requirements enforced at the edge/CDN.

### Anti-Enumeration

Responses for non-existent users, invalid passwords, or suspended accounts are intentionally unified:
*   Standardized HTTP Status: `401 Unauthorized` / `gRPC Unauthenticated`.
*   Standardized Message: `Invalid credentials`.

### CORS & CSRF Protocol

*   **CORS:** Explicitly constrained to authorized front-end client domains defined in configuration. Credentials mode (`Access-Control-Allow-Credentials`) is enabled.
*   **CSRF:** For browser-based REST clients, Access Tokens can be returned in JSON, but if `HttpOnly` cookies are configured for session management, a `X-CSRF-Token` header check is strictly enforced on all state-mutating (`POST`, `PUT`, `DELETE`) endpoints.