package net.minecraftserverlauncher.velocity.config;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Minimal TOML-subset loader for the plugin's velocity.toml config.
 *
 * <p>Only the keys this plugin needs are read; unknown keys are ignored. The
 * config is local and admin-controlled (not attacker-controlled), but values
 * are still sanitized (a URL must be http/https, a token must not be empty).
 */
public final class PluginConfig {

    private final String backendUrl;
    private final String adminToken;

    private PluginConfig(String backendUrl, String adminToken) {
        this.backendUrl = backendUrl;
        this.adminToken = adminToken;
    }

    public String backendUrl() {
        return backendUrl;
    }

    public String adminToken() {
        return adminToken;
    }

    /**
     * Loads config from path. On any error (missing file, invalid URL, empty
     * token) returns null so the caller can refuse to start.
     */
    public static PluginConfig load(Path path) {
        try {
            if (!Files.isRegularFile(path)) {
                return null;
            }
            String text = Files.readString(path);
            String url = valueOf(text, "backend-url");
            String token = valueOf(text, "admin-token");
            if (url == null || token == null || token.isEmpty()) {
                return null;
            }
            if (!url.startsWith("http://") && !url.startsWith("https://")) {
                return null;
            }
            return new PluginConfig(url, token);
        } catch (Exception e) {
            return null;
        }
    }

    private static String valueOf(String text, String key) {
        // key = "value"  (TOML basic string)
        Pattern p = Pattern.compile("(?m)^\\s*" + Pattern.quote(key) + "\\s*=\\s*\"([^\"]*)\"\\s*$");
        Matcher m = p.matcher(text);
        return m.find() ? m.group(1) : null;
    }
}
