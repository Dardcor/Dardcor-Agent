<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=dardcor%20code&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>
<p align="center">Der Open-Source KI-Coding-Agent.</p>
<p align="center">
  <a href="https://dardcor-code.vercel.app/"><img alt="Discord" src="https://dardcor-code.vercel.app/" /></a>
  <a href="https://dardcor-code.vercel.app/"><img alt="npm" src="https://dardcor-code.vercel.app/" /></a>
  <a href="https://dardcor-code.vercel.app/"><img alt="Build status" src="https://dardcor-code.vercel.app/" /></a>
</p>

<p align="center">
  <a href="README.md">English</a> |
  <a href="README.zh.md">简体中文</a> |
  <a href="README.zht.md">繁體中文</a> |
  <a href="README.ko.md">한국어</a> |
  <a href="README.de.md">Deutsch</a> |
  <a href="README.es.md">Español</a> |
  <a href="README.fr.md">Français</a> |
  <a href="README.it.md">Italiano</a> |
  <a href="README.da.md">Dansk</a> |
  <a href="README.ja.md">日本語</a> |
  <a href="README.pl.md">Polski</a> |
  <a href="README.ru.md">Русский</a> |
  <a href="README.bs.md">Bosanski</a> |
  <a href="README.ar.md">العربية</a> |
  <a href="README.no.md">Norsk</a> |
  <a href="README.br.md">Português (Brasil)</a> |
  <a href="README.th.md">ไทย</a> |
  <a href="README.tr.md">Türkçe</a> |
  <a href="README.uk.md">Українська</a> |
  <a href="README.bn.md">বাংলা</a> |
  <a href="README.gr.md">Ελληνικά</a> |
  <a href="README.vi.md">Tiếng Việt</a>
</p>

[![dardcor Terminal UI](packages/web/src/assets/lander/screenshot.png)](https://dardcor-code.vercel.app/)

---

### Installation

```bash
# YOLO
curl -fsSL https://dardcor-code.vercel.app/ | bash

# Paketmanager
npm i -g dardcor-ai@latest        # oder bun/pnpm/yarn
scoop install dardcor             # Windows
choco install dardcor             # Windows
brew install anomalyco/tap/dardcor # macOS und Linux (empfohlen, immer aktuell)
brew install dardcor              # macOS und Linux (offizielle Brew-Formula, seltener aktualisiert)
sudo pacman -S dardcor            # Arch Linux (Stable)
paru -S dardcor-bin               # Arch Linux (Latest from AUR)
mise use -g dardcor               # jedes Betriebssystem
nix run nixpkgs#dardcor           # oder github:anomalyco/dardcor für den neuesten dev-Branch
```

> [!TIP]
> Entferne Versionen älter als 0.1.x vor der Installation.

### Desktop-App (BETA)

dardcor ist auch als Desktop-Anwendung verfügbar. Lade sie direkt von der [Releases-Seite](https://dardcor-code.vercel.app/) oder [dardcor.ai/download](https://dardcor-code.vercel.app/) herunter.

| Plattform             | Download                           |
| --------------------- | ---------------------------------- |
| macOS (Apple Silicon) | `dardcor-desktop-mac-arm64.dmg`   |
| macOS (Intel)         | `dardcor-desktop-mac-x64.dmg`     |
| Windows               | `dardcor-desktop-windows-x64.exe` |
| Linux                 | `.deb`, `.rpm` oder AppImage       |

```bash
# macOS (Homebrew)
brew install --cask dardcor-desktop
# Windows (Scoop)
scoop bucket add extras; scoop install extras/dardcor-desktop
```

#### Installationsverzeichnis

Das Installationsskript beachtet die folgende Prioritätsreihenfolge für den Installationspfad:

1. `$dardcor_INSTALL_DIR` - Benutzerdefiniertes Installationsverzeichnis
2. `$XDG_BIN_DIR` - XDG Base Directory Specification-konformer Pfad
3. `$HOME/bin` - Standard-Binärverzeichnis des Users (falls vorhanden oder erstellbar)
4. `$HOME/.dardcor/bin` - Standard-Fallback

```bash
# Beispiele
dardcor_INSTALL_DIR=/usr/local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
```

### Agents

dardcor enthält zwei eingebaute Agents, zwischen denen du mit der `Tab`-Taste wechseln kannst.

- **build** - Standard-Agent mit vollem Zugriff für Entwicklungsarbeit
- **plan** - Nur-Lese-Agent für Analyse und Code-Exploration
  - Verweigert Datei-Edits standardmäßig
  - Fragt vor dem Ausführen von bash-Befehlen nach
  - Ideal zum Erkunden unbekannter Codebases oder zum Planen von Änderungen

Außerdem ist ein **general**-Subagent für komplexe Suchen und mehrstufige Aufgaben enthalten.
Dieser wird intern genutzt und kann in Nachrichten mit `@general` aufgerufen werden.

Mehr dazu unter [Agents](https://dardcor-code.vercel.app/).

### Dokumentation

Mehr Infos zur Konfiguration von dardcor findest du in unseren [**Docs**](https://dardcor-code.vercel.app/).

### Beitragen

Wenn du zu dardcor beitragen möchtest, lies bitte unsere [Contributing Docs](./CONTRIBUTING.md), bevor du einen Pull Request einreichst.

### Auf dardcor aufbauen

Wenn du an einem Projekt arbeitest, das mit dardcor zusammenhängt und "dardcor" als Teil seines Namens verwendet (z.B. "dardcor-dashboard" oder "dardcor-mobile"), füge bitte einen Hinweis in deine README ein, dass es nicht vom dardcor-Team gebaut wird und nicht in irgendeiner Weise mit uns verbunden ist.

---

**Tritt unserer Community bei** [Discord](https://dardcor-code.vercel.app/) | [X.com](https://dardcor-code.vercel.app/)
