package net.minecraftserverlauncher.velocity;

import com.google.inject.Inject;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.proxy.ProxyInitializeEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.plugin.annotation.DataDirectory;
import com.velocitypowered.api.proxy.ProxyServer;
import java.nio.file.Path;
import org.slf4j.Logger;

@Plugin(
        id = "minecraftserverlauncher",
        name = "MinecraftServerLauncher Auth",
        version = "0.1.0",
        description = "Authorizes backend-authenticated players on the proxy.",
        authors = {"Ddking-1"})
public final class LauncherVelocityPlugin {

    private final ProxyServer server;
    private final Logger logger;
    private final Path dataDirectory;

    @Inject
    public LauncherVelocityPlugin(ProxyServer server, Logger logger,
            @DataDirectory Path dataDirectory) {
        this.server = server;
        this.logger = logger;
        this.dataDirectory = dataDirectory;
    }

    @Subscribe
    public void onProxyInitialization(ProxyInitializeEvent event) {
        // Backend proxy-auth wiring is added in a follow-up commit. For now
        // the plugin initializes cleanly so the Gradle build is verifiable.
        logger.info("LauncherVelocityPlugin initializing; data dir={}", dataDirectory);
    }

    public ProxyServer getServer() {
        return server;
    }
}
