/**
 * Central frontend configuration.
 *
 * Everything that is expected to change between deployments (branding,
 * servers, alpha-only switches) lives here so views stay generic.
 */
export const config = {
  brand: {
    name: 'Alpha Network',
    shortName: 'Alpha',
    tagline: 'Privát Minecraft élmény',
    alphaLabel: 'ALPHA',
  },

  storagePrefix: 'alpha.web.',

  features: {
    // Alpha-only UI that lets staff preview permission combinations.
    // Real authorization always stays on the server (see docs/SECURITY.md).
    permissionSimulator: true,
    skinUpload: true,
  },

  /**
   * Multi-server support is prepared but intentionally fixed to one server
   * for the alpha. Add another object with `enabled: true` and the server
   * switcher appears automatically in the top bar; stats and leaderboards
   * are keyed per server id.
   *
   * A launcher can preselect a server with `?server=<id>` or `?instance=<id>`.
   */
  servers: [
    {
      id: 'alpha-main',
      name: 'Alpha Main',
      region: 'EU · Budapest',
      address: 'play.alpha-network.hu',
      enabled: true,
      online: true,
      players: 42,
      maxPlayers: 100,
    },
    // {
    //   id: 'alpha-farm',
    //   name: 'Alpha Farm',
    //   region: 'EU · Budapest',
    //   address: 'farm.alpha-network.hu',
    //   enabled: true,
    //   online: true,
    //   players: 18,
    //   maxPlayers: 60,
    // },
  ],

  /**
   * The alpha frontend runs on a local mock provider so the UI can be
   * reviewed without the Go backend. The component keeps the same async
   * shape as a real HTTP client; see docs/WEB_FRONTEND.md for the swap.
   */
  dataSource: 'mock',
};

export default config;
