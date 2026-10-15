# Privileged editing

axam does not need root to edit files you already own.

`sudo axam /etc/nginx/nginx.conf` starts the editor as root. Saving uses the same atomic replace as an ordinary user: a temporary file in the destination directory, then a rename. That works when root can create a file in that directory. It is still a long-running root process, which is a poor default.

Prefer `sudoedit`, or `sudo -e`, when the system has it:

```
sudoedit /etc/nginx/nginx.conf
export EDITOR=axam
sudoedit /etc/nginx/nginx.conf
```

`sudoedit` copies the file to a directory you own, runs `$EDITOR` as you, and replaces the original only after the editor exits. axam does not special-case this. If `EDITOR` is `axam`, the copy opens like any other file.

Do not use `sudo axam` merely to follow an example. Use it only when you intend the process to run as root and you accept that the whole session is privileged.
