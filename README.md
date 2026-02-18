# 🐰 TBunny

> A fast, keyboard-driven terminal UI for managing RabbitMQ clusters

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg)

## ✨ What is TBunny?

TBunny is your friendly terminal companion for RabbitMQ management. No more clicking through web interfaces – manage your queues, exchanges, virtual hosts, and users right from your terminal with lightning-fast keyboard shortcuts!

### 🎓 A Learning Journey

This project started as a personal adventure to learn Go, and it's been heavily inspired by the incredible [k9s](https://github.com/derailed/k9s) project. Both the architecture and the visual design are based on k9s with some modifications. If you're familiar with k9s, you'll feel right at home here! Huge thanks to **Fernand Galiana** and the entire k9s team for creating such an amazing tool and sharing their knowledge with the community.

## 🚀 Features

- ⚡ **Lightning Fast** – Navigate RabbitMQ resources with keyboard shortcuts
- 🎯 **Multi-Cluster Support** – Easily switch between different RabbitMQ clusters
- 📊 **Comprehensive Views** – Queues, exchanges, virtual hosts, users, and more
- 🎨 **Customizable** – Tweak the UI to match your preferences

## 📦 Installation

### Homebrew (macOS)

The easiest way to install on macOS:

```bash
brew install anadale/tbunny/tbunny
```

### Download Pre-built Binaries

Download the latest release for your platform from the [**Releases page**](https://github.com/anadale/tbunny/releases/latest).

#### Quick Install Scripts

**macOS:**
```bash
# Detects your architecture automatically
curl -s https://api.github.com/repos/anadale/tbunny/releases/latest \
  | grep "browser_download_url.*Darwin.*tar.gz" \
  | grep "$(uname -m)" \
  | cut -d '"' -f 4 \
  | xargs curl -LO
tar xzf tbunny_*_Darwin_*.tar.gz
sudo mv tbunny /usr/local/bin/
rm tbunny_*_Darwin_*.tar.gz
```

**Linux:**
```bash
# Detects your architecture automatically
curl -s https://api.github.com/repos/anadale/tbunny/releases/latest \
  | grep "browser_download_url.*Linux.*tar.gz" \
  | grep "$(uname -m)" \
  | cut -d '"' -f 4 \
  | xargs curl -LO
tar xzf tbunny_*_Linux_*.tar.gz
sudo mv tbunny /usr/local/bin/
rm tbunny_*_Linux_*.tar.gz
```

**Windows:**

Download the appropriate `.zip` file from the [**Releases page**](https://github.com/anadale/tbunny/releases/latest), extract it, and add `tbunny.exe` to your PATH.

### Install via Go

If you have Go 1.21+ installed:

```bash
go install github.com/anadale/tbunny/cmd/tbunny@latest
```

### Build from Source

```bash
git clone https://github.com/anadale/tbunny.git
cd tbunny
go build -o tbunny ./cmd/tbunny
sudo mv tbunny /usr/local/bin/
```

## 🎮 Getting Started

Launch TBunny and start managing your RabbitMQ clusters:

```bash
tbunny
```

### First Time Setup

When you run TBunny for the first time (or when no cluster is configured), you'll see the clusters list view. To add your first cluster:

1. Press `a` to open the "Add Cluster" dialog
2. Enter your RabbitMQ Management API URL (e.g., `http://localhost:15672`)
3. Provide your username and password
4. Give your cluster a name
5. Press `Enter` to save

That's it! TBunny will connect to your cluster and you can start managing your RabbitMQ resources.

### Command Line Options

Need debugging logs? No problem:

```bash
tbunny --log-file ~/tbunny-debug.log
```

Want to use a custom config location?

```bash
tbunny --config-dir ~/.my-custom-config
```

## ⌨️ Keyboard Shortcuts

### Global Shortcuts

| Shortcut | Action |
|----------|--------|
| `?` | Toggle help screen |
| `Esc` | Go back / Clear |
| `Ctrl+C` | Exit TBunny |
| `Ctrl+E` | Show/hide header |
| `Ctrl+G` | Show/hide breadcrumbs |

### Resource Navigation

Once you're connected to a cluster, use these shortcuts to jump between views:

| Shortcut | View |
|----------|------|
| `Shift+Q` | 📦 Queues |
| `Shift+E` | 🔄 Exchanges |
| `Shift+V` | 🏠 Virtual Hosts |
| `Shift+U` | 👥 Users |
| `Shift+L` | 🌐 Clusters |

## ⚙️ Configuration

TBunny stores its configuration following the XDG Base Directory spec:

| OS | Configuration Path |
|----|-------------------|
| 🐧 Linux | `~/.config/tbunny/` |
| 🍎 macOS | `~/Library/Application Support/tbunny/` |
| 🪟 Windows | `%APPDATA%\tbunny\` |

### Main Config (`config.yaml`)

Customize the UI behavior:

```yaml
ui:
  enableMouse: true      # Enable mouse clicks and scrolling
  splashDuration: 1s     # How long to show the splash screen
```

**Available Options:**

- **`ui.enableMouse`** (boolean)
  Turn mouse support on or off. Default: `true`

- **`ui.splashDuration`** (duration)
  Control splash screen duration. Examples: `1s`, `500ms`, `2s`. Default: `1s`

### Cluster Configuration

Cluster connections are managed through the TBunny interface. Use the clusters view (`Shift+L`) to add, edit, or remove cluster connections. All cluster configurations are automatically saved to the configuration directory.

## 🛠️ Command Line Flags

```
tbunny [flags]

Flags:
  --log-file string      Path to log file for debugging
  --config-dir string    Override default configuration directory
```

## 📄 License

Licensed under the Apache License 2.0. See LICENSE for details.

## 🙏 Acknowledgments

This project wouldn't exist without the inspiration from [k9s](https://github.com/derailed/k9s). A massive thank you to **Fernand Galiana** and all the k9s contributors for creating such an excellent example of a terminal UI done right!

---

Made with 💚 and lots of ☕
