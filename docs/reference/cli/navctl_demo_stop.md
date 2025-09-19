## navctl demo stop

Stop demo Kind clusters

### Synopsis

Stop (delete) demo Kind clusters.

This command deletes the specified Kind cluster(s) and cleans up associated resources.
If --count is specified, it will stop multiple clusters with numbered suffixes.
Otherwise, it will stop the single cluster with the specified name.

```
navctl demo stop [flags]
```

### Options

```
  -h, --help          help for stop
      --name string   Name of the demo cluster(s) (default "navigator-demo")
```

### Options inherited from parent commands

```
      --log-format string   Log format (text, json) (default "text")
      --log-level string    Log level (debug, info, warn, error) (default "info")
```

### SEE ALSO

* [navctl demo](navctl_demo.md)	 - Manage demo Kind clusters for testing Navigator

