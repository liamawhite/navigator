## navctl local

Run manager and edge services locally

### Synopsis

Run both Navigator manager and edge services locally.
This command starts the manager service first, then connects edge services to it.
By default, it uses the current context from your kubeconfig. You can specify
multiple contexts using the --contexts flag to monitor multiple clusters simultaneously.

The --contexts flag supports both exact context names and glob patterns:
  * - matches any sequence of characters
  ? - matches any single character
  [abc] - matches any character in brackets
  [a-z] - matches any character in range

Examples:
  # Use current context
  navctl local

  # Use specific contexts
  navctl local --contexts context1,context2,context3

  # Use glob patterns to select multiple contexts
  navctl local --contexts "*-prod"
  navctl local --contexts "team-*"
  navctl local --contexts "*-prod,*-staging"
  navctl local --contexts "dev-*,test-?"

  # Mix exact names and patterns
  navctl local --contexts "production,*-staging"

  # Use custom kubeconfig with patterns
  navctl local --kube-config ~/.kube/config --contexts "*-prod"

Available contexts will be shown from your kubeconfig file.
```
navctl local [flags]
```

### Options

```
      --contexts strings       Comma-separated list of kubeconfig contexts to use (uses current context if not specified)
      --disable-ui             Disable UI server
  -h, --help                   help for local
  -k, --kube-config string     Path to kubeconfig file (default "~/.kube/config")
      --manager-host string    Host for manager service (default "localhost")
      --manager-port int       Port for manager service (default 8080)
      --max-message-size int   Maximum gRPC message size in MB (default 10)
      --no-browser             Don't open browser automatically
      --ui-port int            Port for UI server (default 8082)
```

### Options inherited from parent commands

```
      --log-format string   Log format (text, json) (default "text")
      --log-level string    Log level (debug, info, warn, error) (default "info")
```

### SEE ALSO

* [navctl](navctl.md)	 - Navigator control plane CLI

