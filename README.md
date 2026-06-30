<div align="center">

# 🔗 skyping

**Share your terminal. Instantly. Peer-to-peer.**

No VPN. No SSH keys. No config files.
A 6-digit code connects your Linux terminal to anyone — right in their browser.

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/platform-linux-blue)](https://jeetkumar.space)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-00ADD8?logo=go)](https://go.dev)

[**Try it now →**](https://jeetkumar.space) · [Report a bug](https://github.com/Jeetkumarsahu/skyping/issues) · [Request a feature](https://github.com/Jeetkumarsahu/skyping/issues)

</div>

---

## Why skyping?

Helping a teammate debug something over a screenshot is painful. Setting up SSH access just for a 5-minute favor is overkill. **skyping** gives you a live, full-featured terminal session in someone's browser — with nothing more than a 6-digit code.

```sh
curl -fsSL https://jeetkumar.space/install.sh | sh
skyping agent
```

That's it. Share the code. They connect from any browser — no install required on their end.

---

## ✨ Features

| | |
|---|---|
| 🔒 **Encrypted relay** | Your terminal data never sits on disk anywhere |
| 🌐 **Works over the internet** | No same-network or VPN requirement |
| 📱 **Mobile-friendly** | Connect from a phone browser if you need to |
| 💻 **Full PTY support** | `vim`, `htop`, `tmux` — all work exactly as they should |
| ⚡ **Zero config** | One binary, one command, no YAML to write |
| 🆓 **Free & open source** | MIT licensed, audit the code yourself |

---

## 🚀 Quick Start

### 1. Install

```sh
curl -fsSL https://jeetkumar.space/install.sh | sh
```

Detects your architecture, installs to `~/.local/bin` or `/usr/local/bin`, optionally sets up a systemd service.

### 2. Share your terminal

```sh
skyping agent
```
Skyping agent running
Your code: 482 916
Share this code with your teammate
### 3. They connect

They open **[jeetkumar.space/connect.html](https://jeetkumar.space/connect.html)**, type in the code, and your terminal appears — live — in their browser.

---

## 🧱 How it works
A lightweight relay server pairs your agent with their browser session using your 6-digit code, then streams raw terminal I/O between them in real time via WebSockets. No session data is persisted.

---

## 🛠️ Supported platforms

| Distro          | Versions      |
|------------------|---------------|
| Ubuntu           | 20.04+        |
| Debian           | 11+           |
| Fedora           | 38+           |
| Arch Linux       | rolling       |
| openSUSE         | Leap 15.4+    |
| Raspberry Pi OS  | Bullseye+     |
| Kali Linux       | rolling       |

> Built in Go — runs anywhere a Go binary runs. Windows support via WSL works too.

---

## 📦 Building from source

Requires Go 1.21+.

```sh
git clone https://github.com/Jeetkumarsahu/skyping
cd skyping
go build -o skyping ./cmd/skyping
```

---

## 🗺️ Roadmap

- [x] Linux agent + browser terminal
- [x] Relay-based connection (works across networks)
- [ ] Direct peer-to-peer (WebRTC) — no relay server needed
- [ ] Windows native support
- [ ] Session recording / playback
- [ ] Multi-user sessions (pair programming mode)

Have an idea? [Open an issue](https://github.com/Jeetkumarsahu/skyping/issues) — contributions welcome!

---

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you'd like to change.

```sh
git clone https://github.com/Jeetkumarsahu/skyping
cd skyping
# make your changes
go build -o skyping ./cmd/skyping
```

---

## 📄 License

MIT © [Jeet Kumar](https://github.com/Jeetkumarsahu)

---

<div align="center">

**[jeetkumar.space](https://jeetkumar.space)** · Built with Go, WebSockets, and way too much chai ☕

If this saved you a headache, consider giving it a ⭐

</div>
