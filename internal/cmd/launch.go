package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/inconshreveable/mousetrap"
	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

// startedByExplorer Windows'ta exe'ye Explorer'dan çift tıklanıp
// tıklanmadığını söyler (diğer işletim sistemlerinde her zaman false).
// Testlerde değiştirilebilsin diye değişken.
var startedByExplorer = mousetrap.StartedByExplorer

// exitFunc os.Exit yerine kullanılır; çift tıklama modunda pencere
// kapanmadan önce kullanıcının sonucu okuyabilmesi için beklenir.
var exitFunc = os.Exit

// autoScanArgs çift tıklanarak (argümansız) açıldıysa çalıştırılacak
// argümanları döner; aksi halde nil.
func autoScanArgs(args []string, explorer bool) []string {
	if explorer && len(args) == 0 {
		return []string{"scan"}
	}
	return nil
}

// waitForEnter pencere kapanmadan önce Enter bekler.
func waitForEnter(in io.Reader, out io.Writer) {
	fmt.Fprint(out, i18n.T("pause.enter"))
	_, _ = bufio.NewReader(in).ReadString('\n')
}

// prepareDoubleClick çift tıklama modunu hazırlar: cobra'nın "This is a
// command line tool" uyarısını kapatır, varsayılan olarak `scan` çalıştırır
// ve programdan çıkmadan önce Enter bekler. Dönen fonksiyon, Execute
// bittikten sonra çağrılmalıdır.
func prepareDoubleClick(root *cobra.Command) func() {
	args := autoScanArgs(os.Args[1:], startedByExplorer())
	if args == nil {
		return func() {}
	}
	cobra.MousetrapHelpText = ""
	root.SetArgs(args)
	pause := func() { waitForEnter(os.Stdin, os.Stderr) }
	exitFunc = func(code int) {
		pause()
		os.Exit(code)
	}
	return pause
}
