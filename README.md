# 🚀 gomc-ping
[![Go Version](https://img.shields.io/github/go-mod/go-version/cubexteam/gomc-ping)](https://golang.org)
[![License](https://img.shields.io/github/license/cubexteam/gomc-ping)](https://github.com/cubexteam/gomc-ping/blob/main/LICENSE)
[![Multi-Game](https://img.shields.io/badge/Multi--Game-Support-orange)](#-supported-games--protocols)

[Русская документация здесь 🇷🇺](README_RU.md)

High-performance multi-game server ping library for Go.
Supports Minecraft (Java/Bedrock), Source Engine (Rust, CS2), Terraria, **FiveM**, and **SA-MP**. Designed for bots and monitoring systems with built-in caching and proxy-safe protocols.

## ✨ Supported Games & Protocols

| Game / Engine | Protocol | Status |
| :--- | :--- | :--- |
| **Minecraft (Java)** | Proxy-safe handshake (Hypixel/Cloudflare compatible) | ✅ Stable |
| **Minecraft (Bedrock)** | RakNet UDP protocol | ✅ Stable |
| **Source Engine** | A2S protocol with Challenge (0x41) support (Rust, CS2) | ✅ Stable |
| **Terraria** | TShock REST API / TCP Fallback | ✅ Stable |
| **FiveM (GTA V)** | HTTP JSON API (/info.json + /players.json) | ✅ Stable |
| **SA-MP (GTA SA)** | SA-MP UDP Query | ✅ Stable |

## 📦 Installation
```bash
go get github.com/cubexteam/gomc-ping
```

## 🚀 Usage Examples

### Generic Ping (Auto-detect)
```go
import "github.com/cubexteam/gomc-ping"

resp, _ := gomcping.Ping("play.hypixel.net", 25565)
fmt.Println(resp.String())
```

### Game-Specific Ping
```go
// GTA V (FiveM)
fivemResp, _ := gomcping.PingFiveM("95.217.143.11", 30120)

// GTA SA (SA-MP)
sampResp, _ := gomcping.PingSAMP("151.80.47.185", 7777)

// Rust / CS2
rustResp, _ := gomcping.PingRust("1.2.3.4", 28015)

// Terraria
tResp, _ := gomcping.PingTerraria("my-world.com", 7777)
```

### Save Minecraft Favicon
```go
resp, _ := gomcping.Ping("mc.funtime.su", 25565)
gomcping.SaveFavicon(resp.Favicon, "server_icon.png")
```

## 🛠 Features
- **Proxy Compatibility:** Special handshake burst to bypass Cloudflare Spectrum.
- **Smart Caching:** 1-minute TTL cache to prevent rate-limiting.
- **Clean Output:** Automatic MOTD cleaning (removes § and & color codes).
- **Parallel Probing:** Probes multiple protocols simultaneously for the fastest response.
- **Zero-Dependency Core:** Lightweight and fast.

Built with ❤️ by [CubexTeam](https://github.com/cubexteam)
