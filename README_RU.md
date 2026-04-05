# 🚀 gomc-ping
[![Go Version](https://img.shields.io/github/go-mod/go-version/cubexteam/gomc-ping)](https://golang.org)
[![License](https://img.shields.io/github/license/cubexteam/gomc-ping)](https://github.com/cubexteam/gomc-ping/blob/main/LICENSE)
[![Multi-Game](https://img.shields.io/badge/Multi--Game-Поддержка-orange)](#-поддерживаемые-игры)

Высокопроизводительная библиотека для пинга игровых серверов на Go.
Поддерживает Minecraft (Java/Bedrock), Source Engine (Rust, CS2), Terraria, **FiveM** и **SA-MP**. Разработана для ботов и систем мониторинга с кешированием и поддержкой прокси-протоколов.

## ✨ Поддерживаемые игры

| Игра / Движок | Протокол | Статус |
| :--- | :--- | :--- |
| **Minecraft (Java)** | Proxy-safe handshake (Hypixel/Cloudflare compatible) | ✅ Стабильно |
| **Minecraft (Bedrock)** | RakNet UDP protocol | ✅ Стабильно |
| **Source Engine** | A2S protocol с поддержкой Challenge (Rust, CS2, DayZ, ARK, Valheim, Unturned) | ✅ Стабильно |
| **Terraria** | TShock REST API / TCP Fallback | ✅ Стабильно |
| **FiveM (GTA V)** | HTTP JSON API (/info.json + /players.json) | ✅ Стабильно |
| **SA-MP (GTA SA)** | SA-MP UDP Query | ✅ Стабильно |

## 📦 Установка
```bash
go get github.com/cubexteam/gomc-ping
```

## 🚀 Примеры использования

### Универсальный пинг (Авто-детект)
```go
import "github.com/cubexteam/gomc-ping"

resp, _ := gomcping.Ping("play.hypixel.net", 25565)
fmt.Println(resp.String())
```

### Специализированный пинг
```go
// GTA V (FiveM)
fivemResp, _ := gomcping.PingFiveM("95.217.143.11", 30120)

// GTA SA (SA-MP)
sampResp, _ := gomcping.PingSAMP("151.80.47.185", 7777)

// DayZ
dayzResp, _ := gomcping.PingDayZ("1.2.3.4", 2302)

// ARK / Valheim / Unturned
arkResp, _ := gomcping.PingARK("1.2.3.4", 27015)
valheimResp, _ := gomcping.PingValheim("1.2.3.4", 2456)
unturnedResp, _ := gomcping.PingUnturned("1.2.3.4", 27015)

// Rust / CS2
rustResp, _ := gomcping.PingRust("1.2.3.4", 28015)

// Terraria
tResp, _ := gomcping.PingTerraria("my-world.com", 7777)
```

### Сохранение Favicon (Minecraft)
```go
resp, _ := gomcping.Ping("mc.funtime.su", 25565)
gomcping.SaveFavicon(resp.Favicon, "server_icon.png")
```

## 🛠 Особенности
- **Совместимость с прокси:** Специальный handshake-burst для обхода Cloudflare Spectrum.
- **Умное кеширование:** 1-минутный TTL кеш для предотвращения лимитов.
- **Чистый вывод:** Автоматическая очистка MOTD (удаление цветовых кодов § и &).
- **Параллельное сканирование:** Одновременный опрос нескольких протоколов для максимально быстрого ответа.
- **Zero-Dependency Core:** Легковесность и высокая скорость работы.

Built with ❤️ by [CubexTeam](https://github.com/cubexteam)
