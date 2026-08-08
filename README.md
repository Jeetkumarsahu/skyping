<div align="center">

# 🔗 skyping

**Share your terminal privately. Instantly.**

No VPN. No SSH keys. No central relay to operate.
A one-time encrypted link connects your Linux terminal to anyone — right in their browser.

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/platform-linux-blue)](https://jeetkumar.space)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-00ADD8?logo=go)](https://go.dev)

[**Try it now →**](https://jeetkumar.space) · [Report a bug](https://github.com/Jeetkumarsahu/skyping/issues) · [Request a feature](https://github.com/Jeetkumarsahu/skyping/issues)

</div>

---

## Why skyping?

Helping a teammate debug something over a screenshot is painful. Setting up SSH access just for a 5-minute favor is overkill. **skyping** gives you a live, full-featured terminal session in someone's browser through a one-time encrypted link.

```sh
curl -fsSL https://jeetkumar.space/install.sh | sh
skyping agent
```

That's it. Share the printed link. They connect from any browser — no install required on their end.

---

## ✨ Features

| | |
|---|---|
| 🔐 **End-to-end encrypted** | Terminal input and output are encrypted before they leave either device |
| 🏠 **Local session host** | The sharer's machine starts the session server; no Skyping relay is required |
| 🌐 **Works over the internet** | A temporary Cloudflare Quick Tunnel reaches the local session through NAT |
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

Detects your architecture and installs both `skyping` and the `cloudflared` tunnel helper to `~/.local/bin` or `/usr/local/bin`.

### 2. Share your terminal

```sh
skyping agent
```

Example output:

```text
Skyping agent running
Share this one-time encrypted link:
https://jeetkumar.space/connect.html#...
```

### 3. They connect

Send the full printed link to the person you trust. They open it in a modern browser and the terminal appears live. The tunnel address, session authentication secret, and ephemeral public key are held in the URL fragment, so they are not sent to GitHub Pages.

> Treat the link as a password: anyone with it can control the shared terminal while the session is running.

---

## 🧱 How it works
The agent starts a WebSocket server on `127.0.0.1`, then exposes only that local port through a temporary Cloudflare Quick Tunnel. The browser uses the link fragment to authenticate with the agent and derive an AES-GCM key through P-256 ECDH. Terminal input and output are encrypted before WebSocket transport; stopping the agent closes the local server and tunnel.

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
