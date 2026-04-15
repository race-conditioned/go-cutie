package cutie

import "os"

var (
	cReset   = "\x1b[0m"
	cBold    = "\x1b[1m"
	cDim     = "\x1b[2m"
	cRed     = "\x1b[31m"
	cYellow  = "\x1b[33m"
	cBlue    = "\x1b[34m"
	cMagenta = "\x1b[35m"
	cCyan    = "\x1b[36m"
	cWhite   = "\x1b[97m"
)

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func levelColor(level Level) string {
	switch level {
	case LevelDebug:
		return cDim
	case LevelInfo:
		return cCyan
	case LevelWarn:
		return cYellow
	case LevelError:
		return cRed
	default:
		return cWhite
	}
}
