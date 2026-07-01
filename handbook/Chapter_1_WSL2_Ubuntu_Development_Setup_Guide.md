# WSL2 + Ubuntu Development Environment Setup

*A reusable setup guide for Linux development on Windows (WSL2),
suitable for the MIT 6.5840/6.824 Distributed Systems labs.*

------------------------------------------------------------------------

# Step 1. Update Ubuntu

Whenever you create a fresh Ubuntu installation, start by updating it.

``` bash
sudo apt update
sudo apt upgrade -y
```

### Why?

-   `apt update` refreshes Ubuntu's package index (like refreshing an
    app store).
-   `apt upgrade` installs the latest versions of already-installed
    software.

------------------------------------------------------------------------

# Step 2. Install Essential Development Tools

``` bash
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    unzip \
    vim \
    tree \
    htop \
    ripgrep \
    jq
```

### What each package does

  Package           Purpose
  ----------------- -----------------------------------
  build-essential   GCC, G++, make, libc headers
  git               Version control
  curl              Download files via HTTP
  wget              Alternative downloader
  unzip             Extract zip archives
  vim               Terminal text editor
  tree              Display directory trees
  htop              Interactive system monitor
  ripgrep           Extremely fast code search (`rg`)
  jq                JSON processor

------------------------------------------------------------------------

# Step 3. Create a Linux Workspace

Keep your source code inside the Linux filesystem instead of `/mnt/c`.

``` bash
mkdir -p ~/projects
cd ~/projects
pwd
```

Expected output:

``` text
/home/<your_username>/projects
```

Recommended layout:

``` text
~/projects
├── mit-6.5840
├── leetcode
├── system-design
├── mini-llm-serving
└── network-science
```

------------------------------------------------------------------------

# Step 4. Install and Configure VS Code

Install Visual Studio Code on Windows.

Install the **WSL** extension by Microsoft.

From Ubuntu:

``` bash
cd ~/projects
code .
```

If everything is configured correctly:

-   VS Code opens automatically.
-   The lower-left corner displays **WSL: Ubuntu**.
-   The integrated terminal is Bash instead of PowerShell.

If prompted, trust the folder so all features are enabled.

------------------------------------------------------------------------

# Step 5. Install Go

For the MIT distributed systems labs, prefer the **official Go
distribution** instead of Ubuntu's package.

Reasons:

-   Keeps pace with official Go releases.
-   Avoids Ubuntu package lag.
-   Matches common professional Go development workflows.

(Install this after the Linux environment is working.)

------------------------------------------------------------------------

# Step 6. Useful Command-Line Tools

You will frequently use:

``` bash
pwd
ls
cd
mkdir
cp
mv
rm
cat
less
grep
find
history
```

Learning these early will make Linux much more comfortable.

------------------------------------------------------------------------

# Step 7. Clone the MIT Repository

After Go is installed:

``` bash
cd ~/projects
git clone <MIT repository URL>
```

Keep the repository under:

``` text
~/projects/mit-6.5840
```

rather than somewhere under `/mnt/c`.

------------------------------------------------------------------------

# Step 8. Learn Just Enough Linux

Before diving into the labs, become comfortable with:

-   Linux filesystem layout (`/`, `/home`, `/usr`, `/etc`, `/var`,
    `/mnt`)
-   Navigation (`pwd`, `ls`, `cd`)
-   File operations (`cp`, `mv`, `rm`, `mkdir`)
-   Searching (`find`, `grep`, `ripgrep`)
-   Viewing files (`cat`, `less`)
-   Command history (`history`)

Understanding these basics will make the MIT labs much smoother.

------------------------------------------------------------------------

# Recommended Learning Roadmap

``` text
WSL2
    ↓
Ubuntu
    ↓
Windows Terminal
    ↓
VS Code + WSL
    ↓
Linux Fundamentals
    ↓
Git
    ↓
Go
    ↓
MIT MapReduce
    ↓
Raft
    ↓
Distributed Key-Value Store
```

------------------------------------------------------------------------

# Best Practices

-   Keep source code in `~/projects`, **not** `/mnt/c`.
-   Use **Windows Terminal** as your primary terminal.
-   Launch projects from Ubuntu using:

``` bash
cd ~/projects/<project-name>
code .
```

-   Develop and compile inside Linux.
-   Let VS Code be the editor while WSL provides the runtime.

Happy hacking!
