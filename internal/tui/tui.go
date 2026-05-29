// Package tui terminal kullanıcı arayüzü yardımcıları sağlar:
// açılış banner'ı ve tarama sırasında canlı ilerleme spinner'ı.
package tui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// IsTTY verilen dosyanın bir terminale bağlı olup olmadığını söyler.
// Pipe/dosya yönlendirmesinde false döner → animasyon kapatılır.
func IsTTY(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// Banner açılış logosunu döndürür (sürüm bilgisiyle).
func Banner(version string) string {
	duck := color.YellowString
	name := color.New(color.FgCyan, color.Bold).Sprint("mr-ducksafe")
	sub := color.New(color.Faint).Sprint("yerel kurulu yazilim CVE tarayici")
	return fmt.Sprintf(`
  %s
 %s   %s %s
  %s   %s
   %s
`,
		duck("__"),
		duck("<(o )___"), name, color.New(color.Faint).Sprint(version),
		duck("( ._> /"), sub,
		duck("`---'"),
	)
}

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Progress canlı, tek satırlık spinner + tamamlanan aşamalar için ✓ satırı.
// enabled=false ise tüm metodlar no-op (TTY olmayan ortamlar için).
type Progress struct {
	w       io.Writer
	enabled bool

	mu   sync.Mutex
	msg  string
	stop chan struct{}
	wg   sync.WaitGroup
}

// NewProgress yeni bir Progress oluşturur. enabled genelde IsTTY(stderr).
func NewProgress(w io.Writer, enabled bool) *Progress {
	return &Progress{w: w, enabled: enabled, stop: make(chan struct{})}
}

// Start spinner animasyon döngüsünü başlatır.
func (p *Progress) Start() {
	if !p.enabled {
		return
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-p.stop:
				return
			case <-t.C:
				p.mu.Lock()
				msg := p.msg
				p.mu.Unlock()
				// \r: satır başına dön, \033[K: satır sonuna kadar temizle
				fmt.Fprintf(p.w, "\r\033[K%s %s", color.CyanString(frames[i%len(frames)]), msg)
				i++
			}
		}
	}()
}

// Update aktif spinner mesajını değiştirir.
func (p *Progress) Update(msg string) {
	p.mu.Lock()
	p.msg = msg
	p.mu.Unlock()
}

// Complete mevcut satırı temizler ve yeşil ✓ ile tamamlanan aşamayı yazar.
func (p *Progress) Complete(line string) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.w, "\r\033[K%s %s\n", color.GreenString("✓"), line)
}

// Stop animasyonu durdurur ve son spinner satırını temizler.
func (p *Progress) Stop() {
	if !p.enabled {
		return
	}
	close(p.stop)
	p.wg.Wait()
	fmt.Fprint(p.w, "\r\033[K")
}
