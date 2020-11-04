package md

import (
	"io"

	"github.com/richardwilkes/toolbox/errs"
)

// Option defines an option for the MarkDown processor.
type Option func(*MarkDown) error

// MaxLineSize adjusts the maximum number of bytes that can comprise a single line within an input file. The default is
// bufio.MaxScanTokenSize.
func MaxLineSize(size int) Option {
	return func(m *MarkDown) error {
		if size < 2 {
			return errs.Newf("MaxLineSize of %d is too small", size)
		}
		m.lineBufferSize = size
		return nil
	}
}

// IncludeProvider is called to provide data for include directives. The default is os.Open.
func IncludeProvider(f func(path string) (io.ReadCloser, error)) Option {
	return func(m *MarkDown) error {
		if f == nil {
			return errs.New("IncludeProvider may not be nil")
		}
		m.includeProvider = f
		return nil
	}
}
