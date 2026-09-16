package cli

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/x/ansi"
)

var progressFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// progressDisplay keeps an interactive terminal informed while an apply is
// running. Non-terminal output remains stable and line-oriented for scripts.
type progressDisplay struct {
	writer      io.Writer
	interactive bool
	width       int

	mu      sync.Mutex
	message string
	frame   int
	stop    chan struct{}
	done    chan struct{}
	stopped atomic.Bool
}

func newProgressDisplay(writer io.Writer, interactive bool, width int) *progressDisplay {
	return &progressDisplay{writer: writer, interactive: interactive, width: width}
}

func (p *progressDisplay) Start(initial string) {
	p.Update(initial)
	if !p.interactive {
		return
	}
	p.stop = make(chan struct{})
	p.done = make(chan struct{})
	p.render()
	go func() {
		defer close(p.done)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.mu.Lock()
				p.frame = (p.frame + 1) % len(progressFrames)
				p.mu.Unlock()
				p.render()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *progressDisplay) Update(message string) {
	p.mu.Lock()
	p.message = message
	if !p.interactive {
		fmt.Fprintln(p.writer, message) //nolint:errcheck
	}
	p.mu.Unlock()
	if p.interactive && p.stop != nil && !p.stopped.Load() {
		p.render()
	}
}

func (p *progressDisplay) Stop() {
	if !p.interactive || p.stop == nil {
		return
	}
	if !p.stopped.CompareAndSwap(false, true) {
		return
	}
	close(p.stop)
	<-p.done
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.writer, "\r\x1b[2K") //nolint:errcheck
}

func (p *progressDisplay) render() {
	p.mu.Lock()
	defer p.mu.Unlock()
	line := progressFrames[p.frame] + " " + p.message
	if p.width > 0 {
		line = ansi.Truncate(line, max(1, p.width-1), "…")
	}
	fmt.Fprintf(p.writer, "\r\x1b[2K%s", strings.ReplaceAll(line, "\n", " ")) //nolint:errcheck
}
