# ssh-key setup

## what is an ssh-key

## create one

## import one

if you already have ssh-key from another system, simply copy the ssh keypair

```bash
mkdir ~/.ssh # create .ssh/ folder if not already exist
cd ~/.ssh
vim id_username # :i to insert, paste private ssh-key in, :wq to save
vim id_username.pub # past public ssh-key in
chmod 600 id_username
chmod 600 id_username.pub
```

## make sure the public key is set via github.com website

under settings > SSH and GPG keys > New SSH key

## config setup

```bash
Host github-username # free naming
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_username
  IdentitiesOnly yes

Host github-work
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_work
  IdentitiesOnly yes

Host github-personal
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_personal
  IdentitiesOnly yes
```

## set ssh-agent auto-load script

this is to load the ssh-agent when the terminal booted up

create a bash script file

```bash
mkdir -p ~/.ssh
vim ~/.ssh/ssh-agent-start.sh
```

paste this in

```bash
#!/usr/bin/env bash

AGENT_ENV="$HOME/.ssh/agent_env"
DEFAULT_KEY="$HOME/.ssh/id_username"

start_agent() {
    echo "Starting ssh-agent..."
    ssh-agent -s > "$AGENT_ENV"
    chmod 600 "$AGENT_ENV"
    source "$AGENT_ENV" >/dev/null

    if [ -f "$DEFAULT_KEY" ]; then
        ssh-add "$DEFAULT_KEY" >/dev/null 2>&1
    fi
}

if [ -f "$AGENT_ENV" ]; then
    source "$AGENT_ENV" >/dev/null
fi

if ! ssh-add -l >/dev/null 2>&1; then
    start_agent
fi
```

make it executeable

```bash
chmod 700 ~/.ssh/ssh-agent-start.sh
```

tell bash to run it automatically

```bash
vim ~/.bashrc # to open bash config script

# paste the following at the bottom of the script
[ -f "$HOME/.ssh/ssh-agent-start.sh" ] && source "$HOME/.ssh/ssh-agent-start.sh"

# reload the terminal
source ~/.bashrc
```

now clone your private repo at will

```bash
git clone git@github-username:luxeave/entropy-course.git
```
