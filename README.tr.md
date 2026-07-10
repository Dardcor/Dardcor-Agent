<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=dardcor%20code&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>
<p align="center">Açık kaynaklı yapay zeka kodlama asistanı.</p>
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

### Kurulum

```bash
# YOLO
curl -fsSL https://dardcor-code.vercel.app/ | bash

# Paket yöneticileri
npm i -g dardcor-ai@latest        # veya bun/pnpm/yarn
scoop install dardcor             # Windows
choco install dardcor             # Windows
brew install anomalyco/tap/dardcor # macOS ve Linux (önerilir, her zaman güncel)
brew install dardcor              # macOS ve Linux (resmi brew formülü, daha az güncellenir)
sudo pacman -S dardcor            # Arch Linux (Stable)
paru -S dardcor-bin               # Arch Linux (Latest from AUR)
mise use -g dardcor               # Tüm işletim sistemleri
nix run nixpkgs#dardcor           # veya en güncel geliştirme dalı için github:anomalyco/dardcor
```

> [!TIP]
> Kurulumdan önce 0.1.x'ten eski sürümleri kaldırın.

### Masaüstü Uygulaması (BETA)

dardcor ayrıca masaüstü uygulaması olarak da mevcuttur. Doğrudan [sürüm sayfasından](https://dardcor-code.vercel.app/) veya [dardcor.ai/download](https://dardcor-code.vercel.app/) adresinden indirebilirsiniz.

| Platform              | İndirme                            |
| --------------------- | ---------------------------------- |
| macOS (Apple Silicon) | `dardcor-desktop-mac-arm64.dmg`   |
| macOS (Intel)         | `dardcor-desktop-mac-x64.dmg`     |
| Windows               | `dardcor-desktop-windows-x64.exe` |
| Linux                 | `.deb`, `.rpm` veya AppImage       |

```bash
# macOS (Homebrew)
brew install --cask dardcor-desktop
# Windows (Scoop)
scoop bucket add extras; scoop install extras/dardcor-desktop
```

#### Kurulum Dizini (Installation Directory)

Kurulum betiği (install script), kurulum yolu (installation path) için aşağıdaki öncelik sırasını takip eder:

1. `$dardcor_INSTALL_DIR` - Özel kurulum dizini
2. `$XDG_BIN_DIR` - XDG Base Directory Specification uyumlu yol
3. `$HOME/bin` - Standart kullanıcı binary dizini (varsa veya oluşturulabiliyorsa)
4. `$HOME/.dardcor/bin` - Varsayılan yedek konum

```bash
# Örnekler
dardcor_INSTALL_DIR=/usr/local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
```

### Ajanlar

dardcor, `Tab` tuşuyla aralarında geçiş yapabileceğiniz iki yerleşik (built-in) ajan içerir.

- **build** - Varsayılan, geliştirme çalışmaları için tam erişimli ajan
- **plan** - Analiz ve kod keşfi için salt okunur ajan
  - Varsayılan olarak dosya düzenlemelerini reddeder
  - Bash komutlarını çalıştırmadan önce izin ister
  - Tanımadığınız kod tabanlarını keşfetmek veya değişiklikleri planlamak için ideal

Ayrıca, karmaşık aramalar ve çok adımlı görevler için bir **genel** alt ajan bulunmaktadır.
Bu dahili olarak kullanılır ve mesajlarda `@general` ile çağrılabilir.

[Ajanlar](https://dardcor-code.vercel.app/) hakkında daha fazla bilgi edinin.

### Dokümantasyon

dardcor'u nasıl yapılandıracağınız hakkında daha fazla bilgi için [**dokümantasyonumuza göz atın**](https://dardcor-code.vercel.app/).

### Katkıda Bulunma

dardcor'a katkıda bulunmak istiyorsanız, lütfen bir pull request göndermeden önce [katkıda bulunma dokümanlarımızı](./CONTRIBUTING.md) okuyun.

### dardcor Üzerine Geliştirme

dardcor ile ilgili bir proje üzerinde çalışıyorsanız ve projenizin adının bir parçası olarak "dardcor" kullanıyorsanız (örneğin, "dardcor-dashboard" veya "dardcor-mobile"), lütfen README dosyanıza projenin dardcor ekibi tarafından geliştirilmediğini ve bizimle hiçbir şekilde bağlantılı olmadığını belirten bir not ekleyin.

---

**Topluluğumuza katılın** [Discord](https://dardcor-code.vercel.app/) | [X.com](https://dardcor-code.vercel.app/)
