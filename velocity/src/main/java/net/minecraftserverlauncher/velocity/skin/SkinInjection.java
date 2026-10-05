package net.minecraftserverlauncher.velocity.skin;

import com.velocitypowered.api.util.GameProfile;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Builds a Velocity GameProfile Property carrying a skin texture, as returned
 * by the backend. The texture JSON and signature are treated as untrusted
 * backend data: they are placed verbatim into a Property (which the proxy
 * will relay to the game), and no parsing into executable content occurs.
 */
public final class SkinInjection {

    private SkinInjection() {
    }

    /**
     * Returns a new GameProfile with a "textures" property appended, or the
     * original profile if either value is null/blank.
     */
    public static GameProfile injectSkin(GameProfile profile, String textureJson, String signature) {
        if (textureJson == null || textureJson.isEmpty() || signature == null || signature.isEmpty()) {
            return profile;
        }
        GameProfile.Property property = new GameProfile.Property("textures", textureJson, signature);
        List<GameProfile.Property> props = new ArrayList<>(profile.getProperties());
        props.add(property);
        return new GameProfile(profile.getId(), profile.getName(), props);
    }
}
