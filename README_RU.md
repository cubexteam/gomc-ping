# 🚀 gomc-ping

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Multi-Game](https://img.shields.io/badge/games-Minecraft%20%7C%20Rust%20%7C%20CS2%20%7C%20Terraria-orange?style=flat-square)](https://github.com/cubexteam/gomc-ping)

**Высокопроизводительная библиотека на Go для пинга игровых серверов.**  
Поддерживает Minecraft (Java/Bedrock), Source Engine (Rust, CS:GO, CS2) и Terraria. Идеально подходит для Telegram/VK ботов и систем мониторинга с высокой нагрузкой.

## ✨ Поддерживаемые игры
*   **Minecraft**: 
    *   **Java Edition**: Обход прокси (Hypixel/Cloudflare), поддержка SRV.
    *   **Bedrock Edition**: Протокол RakNet (UDP).
    *   **Query**: GameSpy4 (плагины, карты, ПО).
    *   **RCON**: Удаленное управление консолью.
*   **Source Engine**: 
    *   **Rust / CS2 / CS:GO / GMod**: Протокол A2S с поддержкой Challenge (0x41).
*   **Terraria**: 
    *   **TShock API**: Авто-определение REST API (название мира, список игроков).
    *   **TCP Fallback**: Проверка доступности порта.

## 📦 Установка
```bash
go get github.com/cubexteam/gomc-ping
```

## 🚀 Примеры использования

### Автоматическое определение игры
```go
import "github.com/cubexteam/gomc-ping"

// Либа сама поймет, какая это игра
resp, _ := gomcping.Ping("play.hypixel.net", 25565)
fmt.Println(resp.String())
```

### Пинг конкретной игры
```go
// Rust
rustResp, _ := gomcping.PingRust("1.2.3.4", 28015)

// CS2
csResp, _ := gomcping.PingCS2("5.6.7.8", 27015)

// Terraria
tResp, _ := gomcping.PingTerraria("my-world.com", 7777)
```

### Сохранение иконки Minecraft
```go
resp, _ := gomcping.Ping("mc.funtime.su", 25565)
// Сохраняет Base64 иконку сразу в PNG файл
gomcping.SaveFavicon(resp.Favicon, "icon.png")
```

## 🛠 Особенности
*   **Совместимость с прокси**: Специальный "пакетный" запрос для прохождения через Cloudflare Spectrum.
*   **Умное кэширование**: Встроенный кэш на 1 минуту для защиты от лимитов (rate-limit).
*   **Чистый вывод**: Автоматическая очистка MOTD от цветовых кодов (§ и &).
*   **Минимум зависимостей**: Легкая и быстрая библиотека.

---
Сделано с ❤️ от [CubexTeam](https://github.com/cubexteam)
