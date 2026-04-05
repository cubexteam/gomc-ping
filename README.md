# 🚀 gomc-ping

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Multi-Game](https://img.shields.io/badge/games-Minecraft%20%7C%20Rust%20%7C%20CS2%20%7C%20Terraria-orange?style=flat-square)](https://github.com/cubexteam/gomc-ping)

[**Русская документация здесь 🇷🇺**](README_RU.md)

**High-performance multi-game server ping library for Go.**  
Supports Minecraft (Java/Bedrock), Source Engine (Rust, CS:GO, CS2), and Terraria (TShock). Designed for bots and monitoring systems with built-in caching and proxy-safe protocols.

## ✨ Supported Games & Protocols
*   **Minecraft**: 
    *   **Java Edition**: Proxy-safe handshake (Hypixel/Cloudflare compatible).
    *   **Bedrock Edition**: RakNet UDP protocol.
    *   **Query**: GameSpy4 (plugins, maps, software).
    *   **RCON**: Remote console management.
*   **Source Engine**: 
    *   **Rust / CS2 / CS:GO / GMod**: A2S protocol with Challenge (0x41) support.
*   **Terraria**: 
    *   **TShock API**: Automatic REST API detection (World name, player list).
    *   **TCP Fallback**: Smart port reachability.

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
// Rust
rustResp, _ := gomcping.PingRust("1.2.3.4", 28015)

// CS2
csResp, _ := gomcping.PingCS2("5.6.7.8", 27015)

// Terraria
tResp, _ := gomcping.PingTerraria("my-world.com", 7777)
```

### Save Minecraft Favicon
```go
resp, _ := gomcping.Ping("mc.funtime.su", 25565)
gomcping.SaveFavicon(resp.Favicon, "server_icon.png")
```

## 🛠 Features
*   **Proxy Compatibility**: Special handshake burst to bypass Cloudflare Spectrum.
*   **Smart Caching**: 1-minute TTL cache to prevent rate-limiting.
*   **Clean Output**: Automatic MOTD cleaning (removes § and & color codes).
*   **Zero-Dependency Core**: Lightweight and fast.

---
Built with ❤️ by [CubexTeam](https://github.com/cubexteam)
