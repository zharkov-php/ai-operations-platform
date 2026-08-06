# API authentication

Web sessions use secure HTTP-only cookies. Mobile clients receive short-lived access and rotating refresh tokens stored with Expo SecureStore. Ingestion uses hashed API keys with an identifying prefix, organization scope, optional project restriction, revocation, and explicit scopes.

```mermaid
sequenceDiagram
  participant Client
  participant API
  participant Store
  Client->>API: Login
  API->>Store: Verify password and lockout
  API-->>Client: Cookie or access + refresh token
  Client->>API: Scoped request
  API->>Store: Verify tenant, role, or API-key scope
```

A production deployment can integrate an external OIDC provider such as Keycloak; the API remains the authorization boundary.
