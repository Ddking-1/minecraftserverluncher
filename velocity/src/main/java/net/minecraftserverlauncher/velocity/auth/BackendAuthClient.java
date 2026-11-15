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
 * is sent in the request body (never a query string).
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

    /**
     * Reports whether the backend considers the access token valid.
     */
    public CompletableFuture<Boolean> validateAccessToken(String accessToken) {
        if (accessToken == null || accessToken.isEmpty()) {
            return CompletableFuture.completedFuture(false);
        }
        String json = "{\"access_token\":\"" + jsonEscape(accessToken) + "\"}";
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/v1/proxy/auth"))
                .timeout(Duration.ofSeconds(5))
                .header("Content-Type", "application/json")
                .header("X-Admin-Token", adminToken)
                .POST(HttpRequest.BodyPublishers.ofString(json, StandardCharsets.UTF_8))
                .build();
        return http.sendAsync(request, HttpResponse.BodyHandlers.ofString())
                .thenApply(resp -> resp.statusCode() == 200 && resp.body().contains("\"status\":\"ok\""))
                .completeOnTimeout(false, 5, TimeUnit.SECONDS);
    }

    /**
     * Escapes a value for embedding in a JSON string literal. The access token
     * originates from client-controlled profile data, so it must never be
     * concatenated into JSON verbatim.
     */
    private static String jsonEscape(String value) {
        StringBuilder out = new StringBuilder(value.length() + 8);
        for (int i = 0; i < value.length(); i++) {
            char c = value.charAt(i);
            switch (c) {
                case '"' -> out.append("\\\"");
                case '\\' -> out.append("\\\\");
                case '\b' -> out.append("\\b");
                case '\f' -> out.append("\\f");
                case '\n' -> out.append("\\n");
                case '\r' -> out.append("\\r");
                case '\t' -> out.append("\\t");
                default -> {
                    if (c < 0x20) {
                        out.append(String.format("\\u%04x", (int) c));
                    } else {
                        out.append(c);
                    }
                }
            }
        }
        return out.toString();
    }

    private static String stripTrailingSlash(String url) {
        if (url != null && url.endsWith("/")) {
            return url.substring(0, url.length() - 1);
        }
        return url;
    }
}
