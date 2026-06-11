Here’s a clean, practical guide to installing **fnm (Fast Node Manager)** depending on your system.

---

## 🔧 1. Install fnm

### ✅ macOS (Homebrew)

```bash
brew install fnm
```

---

### ✅ macOS / Linux (curl installer — recommended)

```bash
curl -fsSL https://fnm.vercel.app/install | bash
```

Then restart your shell or run:

```bash
source ~/.bashrc   # or ~/.zshrc / ~/.profile depending on your shell
```

---

### ✅ Windows (PowerShell via Scoop)

If you use Scoop:

```powershell
scoop install fnm
```

Or via Chocolatey:

```powershell
choco install fnm
```

---

## ⚙️ 2. Configure your shell

Add this to your shell config file:

### For **bash (~/.bashrc)**

```bash
eval "$(fnm env --use-on-cd)"
```

### For **zsh (~/.zshrc)**

```bash
eval "$(fnm env --use-on-cd)"
```

Then reload:

```bash
source ~/.bashrc   # or ~/.zshrc
```

---

## 📦 3. Install Node.js using fnm

```bash
fnm install --lts
```

Set it as default:

```bash
fnm default lts-latest
```

---

## 🔍 4. Verify installation

```bash
fnm --version
node -v
```

---

## ⚡ Optional (auto-switch Node versions per folder)

Create a `.node-version` file:

```bash
echo "lts" > .node-version
```

fnm will automatically switch versions when you `cd` into that folder.

---

If you tell me your OS and shell (bash, zsh, fish, PowerShell), I can tailor this into a copy-paste setup specific to you.
