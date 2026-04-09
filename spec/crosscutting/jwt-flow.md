# JWT Validation Flow Architecture

## Current Architecture (After Fix)

```mermaid
flowchart TD
    A[gRPC Request] --> B[Auth Interceptor]
    B --> C[Extract Bearer Token]
    C --> D{Token Present?}
    D -- No --> E[Return codes.Unauthenticated]
    D -- Yes --> F[TokenService.ValidateAccessToken]
    F --> G{Valid Token?}
    G -- No --> H[Return codes.Unauthenticated]
    G -- Yes --> I[Extract Claims]
    I --> J[Inject Claims into Context]
    J --> K[Call Handler with Context]
    
    style B fill:#e1f5fe
    style F fill:#f3e5f5
    style G fill:#fff3e0
```

## Service Dependencies

```mermaid
flowchart LR
    subgraph "Transport Layer"
        A[gRPC Interceptor]
        B[HTTP Middleware]
    end
    
    subgraph "Domain Services"
        C[TokenService]
        D[AuthService]
        E[Other Services]
    end
    
    subgraph "Repository Layer"
        F[UserRepository]
        G[AuthRepository]
        H[Other Repositories]
    end
    
    A --> C
    B --> C
    D --> C
    D --> F
    D --> G
    E --> F
    E --> G
    E --> H
    
    style C fill:#e8f5e8
    style A fill:#ffeaa7
    style B fill:#ffeaa7
```

## Key Design Decisions

### 1. **TokenService Handles JWT Operations**
- **Single Responsibility**: Only token generation, validation, parsing
- **No Business Logic**: Doesn't know about users, tenants, or auth flows
- **Stateless**: No database calls for token validation

### 2. **AuthService Handles Auth Business Logic**
- **Orchestrates**: Calls TokenService for token operations
- **Business Rules**: Login flows, MFA, session management
- **Stateful**: Database calls for user/tenant validation

### 3. **Interceptor is Transport-Agnostic**
- **Clean Separation**: Only extracts tokens and injects claims
- **Error Mapping**: Translates domain errors to gRPC codes
- **No Dependencies**: Only depends on TokenService interface

## Error Flow Mapping

```mermaid
flowchart TD
    A[TokenService.ValidateAccessToken] --> B{Error Type}
    B -->|ErrInvalidToken| C[codes.Unauthenticated]
    B -->|ErrTokenExpired| C
    B -->|Success| D[Claims Object]
    
    E[Interceptor] --> F{Handler Error}
    F -->|Domain Error| G[MapToGRPCode]
    G --> H[Return gRPC Status]
    
    style C fill:#ffcdd2
    style D fill:#c8e6c9
    style G fill:#fff3e0
```

## Implementation Sequence

1. **Phase 1**: Move `ValidateAccessToken` to `TokenService`
2. **Phase 2**: Update interceptor to use `TokenService`
3. **Phase 3**: Add error mapping helper
4. **Phase 4**: Audit and complete remaining services

## Benefits of This Architecture

- ✅ **Testable**: Each service has clear boundaries
- ✅ **Maintainable**: Single responsibility per service
- ✅ **Scalable**: Easy to add new token types or validation rules
- ✅ **Secure**: Clear separation of concerns reduces attack surface
