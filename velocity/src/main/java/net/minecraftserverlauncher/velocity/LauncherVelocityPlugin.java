package net.minecraftserverlauncher.velocity;

import com.google.inject.Inject;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.connection.PostLoginEvent;
import com.velocitypowered.api.event.proxy.ProxyInitializeEvent;
import com.velocitypowered.api.event.player.PlayerChooseInitialServerEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.plugin.annotation.DataDirectory;
import com.velocitypowered.api.proxy.ProxyServer;
import com.velocitypowered.api.proxy.Player;
import com.velocitypowered.api.util.GameProfile;
import net.kyori.adventure.text.Component;
import net.minecraftserverlauncher.velocity.auth.BackendAuthClient;
import net.minecraftserverlauncher.velocity.config.PluginConfig;
import java.nio.file.Path;
import java.util.Optional;
import org.slf4j.Logger;

/**
 * Velocity plugin main class.
 *
 * <p>Authorization model: the proxy is an enforcement point, not the source of
 * truth. On login it asks the Go backend to validate the player's session
 * token and, if the backend denies it, the player is disconnected. The admin
 * token and backend URL come from the local velocity.toml config.
 */
@Plugin(
        id = "minecraftserverlauncher",
        name = "MinecraftServerLauncher Auth",
        version = "0.1.0",
        description = "Authorizes backend-authenticated players on the proxy.",
        authors = {"Ddking-1"})
public final class LauncherVelocityPlugin {

    private static final String TOKEN_PROPERTY = "minecraftserverlauncher-token";

    private final ProxyServer server;
    private final Logger logger;
    private final Path dataDirectory;
    private BackendAuthClient authClient;

    @Inject
    public LauncherVelocityPlugin(ProxyServer server, Logger logger,
            @DataDirectory Path dataDirectory) {
        this.server = server;
        this.logger = logger;
        this.dataDirectory = dataDirectory;
    }

    @Subscribe
    public void onProxyInitialization(ProxyInitializeEvent event) {
        PluginConfig cfg = PluginConfig.load(dataDirectory.resolve("velocity.toml"));
        if (cfg == null) {
            logger.error("LauncherVelocityPlugin: missing/invalid velocity.toml; refusing to authorize players");
            return;
        }
        this.authClient = new BackendAuthClient(cfg.backendUrl(), cfg.adminToken());
        logger.info("LauncherVelocityPlugin initialized; backend={}", cfg.backendUrl());
    }

    /**
     * On login, validate the player's session token against the backend. The
     * token is taken from a GameProfile property set by the launcher; a
     * missing or backend-denied token disconnects the player.
     */
    @Subscribe
    public void onPostLogin(PostLoginEvent event) {
        if (authClient == null) {
            event.getPlayer().disconnect(Component.text("Not authorized"));
            return;
        }
        Player player = event.getPlayer();
        Optional<String> token = findProperty(player, TOKEN_PROPERTY);
        if (token.isEmpty()) {
            player.disconnect(Component.text("Not authorized"));
            return;
        }
        authClient.validateAccessToken(token.get()).whenComplete((valid, err) -> {
            if (Boolean.TRUE.equals(valid)) {
                logger.debug("Player {} authorized", player.getUsername());
            } else {
                player.disconnect(Component.text("Session not authorized"));
            }
        });
    }

    @Subscribe
    public void onPlayerChooseInitialServer(PlayerChooseInitialServerEvent event) {
        // Server selection is deferred: backend authorization is the gate.
        // (Kept as an explicit listener seam for future policy.)
    }

    private static Optional<String> findProperty(Player player, String name) {
        for (GameProfile.Property p : player.getGameProfile().getProperties()) {
            if (name.equals(p.getName())) {
                return Optional.of(p.getValue());
            }
        }
        return Optional.empty();
    }
}
