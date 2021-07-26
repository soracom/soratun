package internal

import (
	"fmt"
	"runtime"
)

// e.g. `soratun/v1.0.0 (Linux amd64 go1.16.6)`
var UserAgent = fmt.Sprintf("soratun/%s (%s %s %s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
