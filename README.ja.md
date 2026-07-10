<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=dardcor%20code&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>
<p align="center">オープンソースのAIコーディングエージェント。</p>
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

### インストール

```bash
# YOLO
curl -fsSL https://dardcor-code.vercel.app/ | bash

# パッケージマネージャー
npm i -g dardcor-ai@latest        # bun/pnpm/yarn でもOK
scoop install dardcor             # Windows
choco install dardcor             # Windows
brew install anomalyco/tap/dardcor # macOS と Linux（推奨。常に最新）
brew install dardcor              # macOS と Linux（公式 brew formula。更新頻度は低め）
sudo pacman -S dardcor            # Arch Linux (Stable)
paru -S dardcor-bin               # Arch Linux (Latest from AUR)
mise use -g dardcor               # どのOSでも
nix run nixpkgs#dardcor           # または github:anomalyco/dardcor で最新 dev ブランチ
```

> [!TIP]
> インストール前に 0.1.x より古いバージョンを削除してください。

### デスクトップアプリ (BETA)

dardcor はデスクトップアプリとしても利用できます。[releases page](https://dardcor-code.vercel.app/) から直接ダウンロードするか、[dardcor.ai/download](https://dardcor-code.vercel.app/) を利用してください。

| プラットフォーム      | ダウンロード                       |
| --------------------- | ---------------------------------- |
| macOS (Apple Silicon) | `dardcor-desktop-mac-arm64.dmg`   |
| macOS (Intel)         | `dardcor-desktop-mac-x64.dmg`     |
| Windows               | `dardcor-desktop-windows-x64.exe` |
| Linux                 | `.deb`、`.rpm`、または AppImage    |

```bash
# macOS (Homebrew)
brew install --cask dardcor-desktop
# Windows (Scoop)
scoop bucket add extras; scoop install extras/dardcor-desktop
```

#### インストールディレクトリ

インストールスクリプトは、インストール先パスを次の優先順位で決定します。

1. `$dardcor_INSTALL_DIR` - カスタムのインストールディレクトリ
2. `$XDG_BIN_DIR` - XDG Base Directory Specification に準拠したパス
3. `$HOME/bin` - 標準のユーザー用バイナリディレクトリ（存在する場合、または作成できる場合）
4. `$HOME/.dardcor/bin` - デフォルトのフォールバック

```bash
# 例
dardcor_INSTALL_DIR=/usr/local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://dardcor-code.vercel.app/ | bash
```

### Agents

dardcor には組み込みの Agent が2つあり、`Tab` キーで切り替えられます。

- **build** - デフォルト。開発向けのフルアクセス Agent
- **plan** - 分析とコード探索向けの読み取り専用 Agent
  - デフォルトでファイル編集を拒否
  - bash コマンド実行前に確認
  - 未知のコードベース探索や変更計画に最適

また、複雑な検索やマルチステップのタスク向けに **general** サブ Agent も含まれています。
内部的に使用されており、メッセージで `@general` と入力して呼び出せます。

[agents](https://dardcor-code.vercel.app/) の詳細はこちら。

### ドキュメント

dardcor の設定については [**ドキュメント**](https://dardcor-code.vercel.app/) を参照してください。

### コントリビュート

dardcor に貢献したい場合は、Pull Request を送る前に [contributing docs](./CONTRIBUTING.md) を読んでください。

### dardcor の上に構築する

dardcor に関連するプロジェクトで、名前に "dardcor"（例: "dardcor-dashboard" や "dardcor-mobile"）を含める場合は、そのプロジェクトが dardcor チームによって作られたものではなく、いかなる形でも関係がないことを README に明記してください。

---

**コミュニティに参加** [Discord](https://dardcor-code.vercel.app/) | [X.com](https://dardcor-code.vercel.app/)
