package net.minecraftserverlauncher.velocity;

import com.google.inject.Inject;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.connection.PostLoginEvent;
import com.velocitypowered.api.event.proxy.ProxyInitializeEvent;
import com.velocitypowered.api.event.player.PlayerChooseInitialServerEvent;
import com.velocitypowered.api.event.player.ServerPreConnectEvent;
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
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;

/**
 * Velocity plugin main class.
 *
 * <p>Authorization model: the proxy is an enforcement point, not the source of
 * truth. On login it asks the Go backend to validate the player's session
 * token and, if the backend denies it, the player is disconnected. A join to
 * any backend server is additionally gated on the player having passed that
 * check, so a player can never reach a backend before authorization. The
 * admin token and backend URL come from the local velocity.toml config.
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

    // Players that have passed backend authorization. Guarded by a concurrent
    // set so the login thread and the connection thread never race.
    private final Set<UUID> authorized = ConcurrentHashMap.newKeySet();

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
     * missing or backend-denied token disconnects the player before they can
     * reach any backend server.
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
        BackendAuthClient.AuthResult result;
        try {
            result = authClient.validateAccessToken(token.get()).get(5, TimeUnit.SECONDS);
        } catch (Exception e) {
            logger.warn("Player {} authorization failed: {}", player.getUsername(), e.toString());
            result = new BackendAuthClient.AuthResult(false, null);
        }
        if (!result.valid()) {
            // A ban reason comes from the admin ban record so the player sees
            // the same text the admin entered; otherwise stay generic.
            String reason = result.reason();
            player.disconnect(Component.text(
                    reason != null && !reason.isBlank() ? "Banned: " + reason : "Session not authorized"));
            return;
        }
        authorized.add(player.getUniqueId());
        logger.debug("Player {} authorized", player.getUsername());
    }

    @Subscribe
    public void onServerPreConnect(ServerPreConnectEvent event) {
        // Reject any join attempt by a player that has not passed backend
        // authorization, regardless of how the connection was initiated.
        if (!authorized.contains(event.getPlayer().getUniqueId())) {
            event.setResult(ServerPreConnectEvent.ServerResult.denied());
        }
    }

    @Subscribe
    public void onPlayerChooseInitialServer(PlayerChooseInitialServerEvent event) {
        // Server selection is deferred: backend authorization is the gate.
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