# Factions for Dragonfly
<p align="center">
  <img src="GoFaction.png" alt="GoFactions Logo" width="220"/>
</p>

![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)

A faction system for the [Dragonfly](https://github.com/df-mc/dragonfly) server software, written entirely in Go.

This project serves as a ready-to-use Factions server, built from the ground up to be robust, scalable, and efficient.

**It is still in development, more things need to be added**

## Features

-   **Faction Management**: Create, delete, and view faction information.
-   **Member System**: Invite players, leave factions.
-   **Permission System**: Leader-only permissions for critical actions like claiming and deleting.
-   **Alliances**: Form alliances with other factions (configurable limit, default is 1). Timed alliance requests (5 minutes).
-   **Chunk-Based Claiming**: Factions can claim chunks of the world.
-   **Territory Borders**: A visual, performance-optimized particle system (`/f border`) to show claimed land.
    -   <span style="color:red">Red</span>: Claimed by another faction.
    -   <span style="color:green">Green</span>: Wilderness (available to claim).
-   **Power System**:
    -   Individual player power.
    -   Faction power calculated from its members.
-   **Leaderboards**: View top players and factions by power (`/f top`).
-   **Data Persistence**: Uses SQLite (`factions.db`) to save all factions, player data, and claims.
-   **Optimized**: Built with a high-performance in-memory cache for fast access to data, and an efficient global task for particle effects to minimize server lag.

## Available Commands

| Command                    | Description                                            |
|----------------------------| ------------------------------------------------------ |
| `/f create <name>`         | Creates a new faction.                                 |
| `/f delete <name>`         | Deletes your faction (leader only, requires name confirmation). |
| `/f info <faction>`        | Shows information about your faction or another one.   |
| `/f top <player\|faction>` | Displays the top 10 players or factions by power.    |
| `/f invite <player>`       | Invites a player to your faction (leader only).        |
| `/f join <faction>`        | Accepts an invitation to a faction.                    |
| `/f leave`                 | Leaves your current faction (cannot be used by leader).|
| `/f ally send <faction>`   | Sends an alliance request (leader only).               |
| `/f ally accept <faction>` | Accepts an alliance request (leader only).             |
| `/f ally deny <faction>`   | Denies an alliance request (leader only).              |
| `/f claim`                 | Claims the chunk you are standing in (leader only).    |
| `/f unclaim`               | Unclaims the chunk you are standing in (leader only).  |
| `/f border`                | Toggles the visual territory borders.                  |

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the GNU License.

---

## 📞 Contact

Need help or have suggestions?

[![Discord: JorgeByte](https://lanyard.cnrad.dev/api/1165097093480853634?theme=dark&bg=005cff&animated=false&hideDiscrim=true&borderRadius=30px&idleMessage=Hello)](https://discord.com/users/1165097093480853634)

[[![Discord: AustinKarasu](
https://lanyard.cnrad.dev/api/927640119723323393?theme=dark&bg=005cff&animated=false&hideDiscrim=true&borderRadius=30px&idleMessage=Hello)
