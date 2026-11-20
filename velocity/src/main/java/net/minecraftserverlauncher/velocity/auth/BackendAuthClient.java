package net.minecraftserverlauncher.velocity.auth;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

/**
 * Small HTTP client to the Go backend proxy-auth endpoint.
 *
 * <p>Security: the admin token is required for the request, the client has a
 * short connect/read timeout and a bounded completion timeout, and the token
 * is sent in the request body (never a query string). The access token is
 * JSON-escaped before being placed in the body so a crafted token cannot
 * break out of the JSON string.
 */
public class BackendAuthClient {

    private final HttpClient http;
    private final String baseUrl;
    private final String adminToken;

    public BackendAuthClient(String baseUrl, String adminToken) {
        this.baseUrl = stripTrailingSlash(baseUrl);
        this.adminToken = adminToken;
        this.http = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(3))
                .build();
    }

    /** Result of a backend authorization check. reason is null when allowed. */
    public record AuthResult(boolean valid, String reason) {
    }

    /**
     * Reports whether the backend considers the access token valid, and the
     * ban reason when the backend denied it so the player can see why.
     */
    public CompletableFuture<AuthResult> validateAccessToken(String accessToken) {
        if (accessToken == null || accessToken.isEmpty()) {
            return CompletableFuture.completedFuture(new AuthResult(false, null));
        }
        String json = "{\"access_token\":\"" + escapeJson(accessToken) + "\"}";
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/v1/proxy/auth"))
                .timeout(Duration.ofSeconds(5))
                .header("Content-Type", "application/json")
                .header("X-Admin-Token", adminToken)
                .POST(HttpRequest.BodyPublishers.ofString(json, StandardCharsets.UTF_8))
                .build();
        return http.sendAsync(request, HttpResponse.BodyHandlers.ofString())
                .thenApply(resp -> {
                    if (resp.statusCode() != 200 || resp.body() == null) {
                        return new AuthResult(false, null);
                    }
                    if (resp.body().contains("\"status\":\"ok\"")) {
                        return new AuthResult(true, null);
                    }
                    return new AuthResult(false, jsonString(resp.body(), "reason"));
                })
                .exceptionally(err -> new AuthResult(false, null))
                .completeOnTimeout(new AuthResult(false, null), 5, TimeUnit.SECONDS);
    }

    /**
     * Reads a string field from the small JSON object this backend returns.
     * Kept dependency-free on purpose; the endpoint's schema is fixed.
     */
    private static String jsonString(String body, String field) {
        String key = "\"" + field + "\":\"";
        int start = body.indexOf(key);
        if (start < 0) {
            return null;
        }
        StringBuilder sb = new StringBuilder();
        boolean escaped = false;
        for (int i = start + key.length(); i < body.length(); i++) {
            char c = body.charAt(i);
            if (escaped) {
                switch (c) {
                    case 'n' -> sb.append('\n');
                    case 'r' -> sb.append('\r');
                    case 't' -> sb.append('\t');
                    case '"', '\\', '/' -> sb.append(c);
                    case 'u' -> {
                        if (i + 4 < body.length()) {
                            try {
                                sb.append((char) Integer.parseInt(body.substring(i + 1, i + 5), 16));
                                i += 4;
                            } catch (NumberFormatException ignored) {
                                // Malformed escape: keep the raw text.
                            }
                        }
                    }
                    default -> sb.append(c);
                }
                escaped = false;
            } else if (c == '\\') {
                escaped = true;
            } else if (c == '"') {
                break;
            } else {
                sb.append(c);
            }
        }
        return sb.toString();
    }

    private static String escapeJson(String s) {
        StringBuilder sb = new StringBuilder(s.length() + 8);
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':
                    sb.append("\\\"");
                    break;
                case '\\':
                    sb.append("\\\\");
                    break;
                case '\n':
                    sb.append("\\n");
                    break;
                case '\r':
                    sb.append("\\r");
                    break;
                case '\t':
                    sb.append("\\t");
                    break;
                default:
                    if (c < 0x20) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
            }
        }
        return sb.toString();
    }

    private static String stripTrailingSlash(String url) {
        if (url != null && url.endsWith("/")) {
            return url.substring(0, url.length() - 1);
        }
        return url;
    }
}