package output

import (
	"regexp"
	"strings"
)

// Redaction patterns for sensitive information
var (
	// Password patterns in JSON/YAML
	passwordPattern = regexp.MustCompile(`("password"\s*:\s*)"[^"]*"`)
	tokenPattern    = regexp.MustCompile(`("token"\s*:\s*)"[^"]*"`)
	authPattern     = regexp.MustCompile(`("auth"\s*:\s*)"[^"]*"`)
	secretPattern   = regexp.MustCompile(`("secret"\s*:\s*)"[^"]*"`)

	// Basic auth in URLs
	basicAuthURLPattern = regexp.MustCompile(`://([^:]+):([^@]+)@`)

	// Bearer tokens
	bearerTokenPattern = regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-._~+/]+=*`)

	// Base64 encoded credentials (common in dockerconfigjson)
	base64CredPattern = regexp.MustCompile(`"auth"\s*:\s*"[A-Za-z0-9+/=]+"`)

	// API keys and tokens (generic)
	apiKeyPattern = regexp.MustCompile(`([Aa]pi[-_]?[Kk]ey|[Tt]oken)\s*[:=]\s*['"]?[A-Za-z0-9\-._~+/]{16,}['"]?`)
)

// RedactSecrets removes sensitive data from strings
func RedactSecrets(input string) string {
	if input == "" {
		return input
	}

	result := input

	// Redact password fields
	result = passwordPattern.ReplaceAllString(result, `$1"[REDACTED]"`)

	// Redact token fields
	result = tokenPattern.ReplaceAllString(result, `$1"[REDACTED]"`)

	// Redact auth fields
	result = authPattern.ReplaceAllString(result, `$1"[REDACTED]"`)

	// Redact secret fields
	result = secretPattern.ReplaceAllString(result, `$1"[REDACTED]"`)

	// Redact basic auth in URLs
	result = basicAuthURLPattern.ReplaceAllString(result, "://[REDACTED]@")

	// Redact bearer tokens
	result = bearerTokenPattern.ReplaceAllString(result, "Bearer [REDACTED]")

	// Redact base64 encoded credentials
	result = base64CredPattern.ReplaceAllString(result, `"auth": "[REDACTED]"`)

	// Redact API keys and tokens
	result = apiKeyPattern.ReplaceAllString(result, "$1: [REDACTED]")

	return result
}

// RedactEventMessage redacts secrets from Kubernetes event messages
func RedactEventMessage(message string) string {
	if message == "" {
		return message
	}

	// Apply general secret redaction
	result := RedactSecrets(message)

	// Additional Kubernetes-specific patterns
	result = redactDockerConfigJSON(result)
	result = redactPullSecrets(result)

	return result
}

// redactDockerConfigJSON redacts docker config credentials
func redactDockerConfigJSON(message string) string {
	// Look for dockerconfigjson patterns
	if strings.Contains(message, "dockerconfigjson") || strings.Contains(message, ".dockercfg") {
		// Redact the entire config if it appears in the message
		dockerConfigPattern := regexp.MustCompile(`\{[^}]*"auth"\s*:\s*"[^"]*"[^}]*\}`)
		message = dockerConfigPattern.ReplaceAllString(message, "{...dockerconfig redacted...}")
	}
	return message
}

// redactPullSecrets redacts references to pull secret content
func redactPullSecrets(message string) string {
	// If message contains references to pull secrets with credentials
	if strings.Contains(message, "imagePullSecrets") || strings.Contains(message, "pull secret") {
		// Redact any credential-like strings that might leak
		credLikePattern := regexp.MustCompile(`[A-Za-z0-9+/]{40,}={0,2}`)
		message = credLikePattern.ReplaceAllString(message, "[REDACTED]")
	}
	return message
}

// RedactSensitiveFields redacts sensitive fields from a map
func RedactSensitiveFields(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	result := make(map[string]interface{})
	sensitiveKeys := []string{
		"password", "token", "auth", "secret",
		"apiKey", "api_key", "apikey",
		"credentials", "credential",
	}

	for key, value := range data {
		lowerKey := strings.ToLower(key)
		isSensitive := false

		for _, sensKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensKey) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			result[key] = "[REDACTED]"
		} else {
			// Recursively redact nested maps
			if nestedMap, ok := value.(map[string]interface{}); ok {
				result[key] = RedactSensitiveFields(nestedMap)
			} else if nestedStr, ok := value.(string); ok {
				result[key] = RedactSecrets(nestedStr)
			} else {
				result[key] = value
			}
		}
	}

	return result
}
