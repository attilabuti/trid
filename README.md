# Trid

Trid is a Go package that provides an interface for the TrID file identifier tool.

## Installation

To use this package, you need to have Go installed on your system. You also need to have the [TrID command-line tool](https://mark0.net/soft-trid-e.html) installed and accessible in your system's PATH.

1. First, install the TrID command-line tool. You can download it from the [official TrID website](https://mark0.net/soft-trid-e.html).

2. Install the Trid Go package:

```bash
$ go get github.com/attilabuti/trid@latest
```

## Usage

Here's a basic example of how to use the trid package:

```go
package main

import (
    "fmt"
    "log"

    "github.com/attilabuti/trid"
)

func main() {
    // Create a new Trid instance with default options
    t := trid.NewTrid(trid.Options{})

    // Scan a file
    fileTypes, err := t.Scan("/path/to/your/file", 3)
    if err != nil {
        log.Fatalf("Error scanning file: %v", err)
    }

    // Print the results
    for _, ft := range fileTypes {
        fmt.Printf("Extension: %s\n", ft.Extension)
        fmt.Printf("Probability: %.2f%%\n", ft.Probability)
        fmt.Printf("Name: %s\n", ft.Name)
        fmt.Printf("MIME Type: %s\n", ft.MimeType)
        fmt.Printf("Related URL: %s\n", ft.RelatedURL)
        fmt.Printf("Definition: %s\n", ft.Definition)
        fmt.Printf("Remarks: %s\n\n", ft.Remarks)
    }
}
```

## Options

You can configure the Trid instance by providing options:

```go
t := trid.NewTrid(trid.Options{
    Cmd:         "/path/to/trid",         // Command to invoke TrID (default: "trid")
    Definitions: "/path/to/triddefs.trd", // Path to TrID definitions file (default: "")
    Timeout:     60 * time.Second,        // Maximum duration to wait for TrID execution (default: 30 * time.Second)
})
```

## Issues

Submit the [issues](https://github.com/attilabuti/trid/issues) if you find any bug or have any suggestion.

## Contribution

Fork the [repo](https://github.com/attilabuti/trid) and submit pull requests.

## License

This extension is licensed under the [MIT License](https://github.com/attilabuti/trid/blob/main/LICENSE).