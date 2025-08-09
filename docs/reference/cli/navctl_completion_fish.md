## navctl completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	navctl completion fish | source

To load completions for every new session, execute once:

	navctl completion fish > ~/.config/fish/completions/navctl.fish

You will need to start a new shell for this setup to take effect.


```
navctl completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --log-format string   Log format (text, json) (default "text")
      --log-level string    Log level (debug, info, warn, error) (default "info")
```

### SEE ALSO

* [navctl completion](navctl_completion.md)	 - Generate the autocompletion script for the specified shell

