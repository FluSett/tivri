# Workspace Customizations & Agent Guidelines

To ensure all future interactions are perfect, safe, and highly collaborative, all AI agents operating in this workspace must adhere to the following rules:

## 1. Safety & Security First
- **No Predictable Credentials/Tokens**: Never write code that generates deterministic session tokens (e.g., static hashes of static credentials). Always use cryptographically secure random session tokens (e.g. from `crypto/rand` or strong UUIDs).
- **Secure Cookie Flags**: Ensure HTTP cookies have the `HttpOnly`, `Secure` (where appropriate), and `SameSite` (e.g., Lax/Strict) attributes.
- **Robust Input Sanitization**: Do not write naive security validation (such as simple substring checking like `strings.Contains` for script tags). Instead, use robust, proven context-aware template escaping or dedicated sanitization libraries.
- **Zero comments policy**: Strictly comply with the comment ban (except compiler directives like `//go:embed` and standard toolchain auto-generated comments like `// indirect` in `go.mod`). No comments, draft instructions, or shell comments are allowed.

## 2. Resource Management & Optimization
- **Database Limits**: Configure explicit connection pool limits (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`) to prevent resource exhaustion.
- **Memory Safety**: Implement TTL eviction or background cleanups for in-memory caching/lockout tracking maps.
- **Exhaustive Error Handling**: Never discard errors using `_` or `_, _`. Every error must be checked and handled.

## 3. Mandatory Collaborative Design Review
- Before performing any code modification or running execution phases:
  1. Share the proposed design, safety implications, and alternatives with the user.
  2. Discuss the technical options and build understanding together.
  3. Obtain explicit user confirmation before modifying the codebase.
## 4. English Language Assistance
- At the end of every response, you must include a dedicated **"Mini Extra Lesson"** section.
- Use this section to correct any grammatical mistakes made by the user in their previous message, suggest more natural/native vocabulary, or teach a quick idiomatic phrasing based on the conversation context.
- **Exception**: Do not correct the user for starting their initial message with a lowercase letter (e.g. "so, did we..."), as this is standard informal chat behavior. However, continue to check and correct capitalization errors for sentences that start after a period/dot.
- Keep the tone helpful, encouraging, and clear.
