<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=dardcor%20code&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>
<p align="center">AI-kodeagent med åpen kildekode.</p>
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

### Installasjon

```bash
# YOLO
curl -fsSL https://dardcor-code.vercel.app/ | bash

# Pakkehåndterere
npm i -g dardcor-ai@latest        # eller bun/pnpm/yarn
scoop install dardcor             # Windows
choco install dardcor             # Windows
brew install anomalyco/tap/dardcor # macOS og Linux (anbefalt, alltid oppdatert)
brew install dardcor              # macOS og Linux (offisiell brew-formel, oppdateres sjeldnere)
sudo pacman -S dardcor            # Arch Linux (Stable)
paru -S dardcor-bin               # Arch Linux (Latest from AUR)
mise use -g dardcor               # alle OS
nix run nixpkgs#dardcor           # eller github:anomalyco/dardcor for nyeste dev-branch
```

> [!TIP]
> Fjern versjoner eldre enn 0.1.x før du installerer.

### Desktop-app (BETA)

dardcor er også tilgjengelig som en desktop-app. Last ned direkte fra [releases-siden](https://dardcor-code.vercel.app/) eller [dardcor.ai/download](https://dardcor-code.vercel.app/).

| Plattform             | Nedlasting                         |
| --------------------- | ---------------------------------- |
| macOS (Apple Silicon) | `dardcor-desktop-mac-arm64.dmg`   |
| macOS (Intel)         | `dardcor-desktop-mac-x64.dmg`     |
| Windows               | `dardcor-desktop-windows-x64.exe` |
| Linux                 | `.deb`, `.rpm` eller AppImage      |

```bash
# macOS (Homebrew)
brew install --cask dardcor-desktop
# Windows (Scoop)
scoop bucket add extras; scoop install extras/dardcor-desktop
```

#### Installasjonsmappe

Installasjonsskriptet bruker følgende prioritet for installasjonsstien:

1. `$dardcor_INSTALL_DIR` - Egendefinert installasjonsmappe
2. `$XDG_BIN_DIR` - Sti som følger XDG Base Directory Specification
3. `$HOME/bin` - Standard brukerbinar-mappe (hvis den finnes eller kan opprettes)
4. `$HOME/.dardcor/bin` - Standard fallback

```bash
# Eksempler
dardcor_INSTALL_DIR=/usr/local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
```

### Agents

dardcor har to innebygde agents du kan bytte mellom med `Tab`-tasten.

- **build** - Standard, agent med full tilgang for utviklingsarbeid
- **plan** - Skrivebeskyttet agent for analyse og kodeutforsking
  - Nekter filendringer som standard
  - Spør om tillatelse før bash-kommandoer
  - Ideell for å utforske ukjente kodebaser eller planlegge endringer

Det finnes også en **general**-subagent for komplekse søk og flertrinnsoppgaver.
Den brukes internt og kan kalles via `@general` i meldinger.

Les mer om [agents](https://dardcor-code.vercel.app/).

### Dokumentasjon

For mer info om hvordan du konfigurerer dardcor, [**se dokumentasjonen**](https://dardcor-code.vercel.app/).

### Bidra

Hvis du vil bidra til dardcor, les [contributing docs](./CONTRIBUTING.md) før du sender en pull request.

### Bygge på dardcor

Hvis du jobber med et prosjekt som er relatert til dardcor og bruker "dardcor" som en del av navnet; for eksempel "dardcor-dashboard" eller "dardcor-mobile", legg inn en merknad i README som presiserer at det ikke er bygget av dardcor-teamet og ikke er tilknyttet oss på noen måte.

---

**Bli med i fellesskapet** [Discord](https://dardcor-code.vercel.app/) | [X.com](https://dardcor-code.vercel.app/)
