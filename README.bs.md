<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=dardcor%20code&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>
<p align="center">dardcor je open source AI agent za programiranje.</p>
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

### Instalacija

```bash
# YOLO
curl -fsSL https://dardcor-code.vercel.app/ | bash

# Package manageri
npm i -g dardcor-ai@latest        # ili bun/pnpm/yarn
scoop install dardcor             # Windows
choco install dardcor             # Windows
brew install anomalyco/tap/dardcor # macOS i Linux (preporučeno, uvijek ažurno)
brew install dardcor              # macOS i Linux (zvanična brew formula, rjeđe se ažurira)
sudo pacman -S dardcor            # Arch Linux (Stable)
paru -S dardcor-bin               # Arch Linux (Latest from AUR)
mise use -g dardcor               # Bilo koji OS
nix run nixpkgs#dardcor           # ili github:anomalyco/dardcor za najnoviji dev branch
```

> [!TIP]
> Ukloni verzije starije od 0.1.x prije instalacije.

### Desktop aplikacija (BETA)

dardcor je dostupan i kao desktop aplikacija. Preuzmi je direktno sa [stranice izdanja](https://dardcor-code.vercel.app/) ili sa [dardcor.ai/download](https://dardcor-code.vercel.app/).

| Platforma             | Preuzimanje                        |
| --------------------- | ---------------------------------- |
| macOS (Apple Silicon) | `dardcor-desktop-mac-arm64.dmg`   |
| macOS (Intel)         | `dardcor-desktop-mac-x64.dmg`     |
| Windows               | `dardcor-desktop-windows-x64.exe` |
| Linux                 | `.deb`, `.rpm`, ili AppImage       |

```bash
# macOS (Homebrew)
brew install --cask dardcor-desktop
# Windows (Scoop)
scoop bucket add extras; scoop install extras/dardcor-desktop
```

#### Instalacijski direktorij

Instalacijska skripta koristi sljedeći redoslijed prioriteta za putanju instalacije:

1. `$dardcor_INSTALL_DIR` - Prilagođeni instalacijski direktorij
2. `$XDG_BIN_DIR` - Putanja usklađena sa XDG Base Directory specifikacijom
3. `$HOME/bin` - Standardni korisnički bin direktorij (ako postoji ili se može kreirati)
4. `$HOME/.dardcor/bin` - Podrazumijevana rezervna lokacija

```bash
# Primjeri
dardcor_INSTALL_DIR=/usr/local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
```

### Agenti

dardcor uključuje dva ugrađena agenta između kojih možeš prebacivati tasterom `Tab`.

- **build** - Podrazumijevani agent sa punim pristupom za razvoj
- **plan** - Agent samo za čitanje za analizu i istraživanje koda
  - Podrazumijevano zabranjuje izmjene datoteka
  - Traži dozvolu prije pokretanja bash komandi
  - Idealan za istraživanje nepoznatih codebase-ova ili planiranje izmjena

Uključen je i **general** pod-agent za složene pretrage i višekoračne zadatke.
Koristi se interno i može se pozvati pomoću `@general` u porukama.

Saznaj više o [agentima](https://dardcor-code.vercel.app/).

### Dokumentacija

Za više informacija o konfiguraciji dardcor-a, [**pogledaj dokumentaciju**](https://dardcor-code.vercel.app/).

### Doprinosi

Ako želiš doprinositi dardcor-u, pročitaj [upute za doprinošenje](./CONTRIBUTING.md) prije slanja pull requesta.

### Gradnja na dardcor-u

Ako radiš na projektu koji je povezan s dardcor-om i koristi "dardcor" kao dio naziva, npr. "dardcor-dashboard" ili "dardcor-mobile", dodaj napomenu u svoj README da projekat nije napravio dardcor tim i da nije povezan s nama.

---

**Pridruži se našoj zajednici** [Discord](https://dardcor-code.vercel.app/) | [X.com](https://dardcor-code.vercel.app/)
