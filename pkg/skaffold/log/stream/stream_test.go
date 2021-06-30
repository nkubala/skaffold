package stream

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/output"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestPrintLogLine(t *testing.T) {
	testutil.Run(t, "verify lines are not intermixed", func(t *testutil.T) {
		var (
			buf  bytes.Buffer
			wg   sync.WaitGroup
			lock sync.Mutex

			linesPerGroup = 100
			groups        = 5
		)

		for i := 0; i < groups; i++ {
			wg.Add(1)

			go func() {
				for i := 0; i < linesPerGroup; i++ {
					printLogLine(output.Default, &buf, func() bool { return false }, &lock, "PREFIX", "TEXT\n")
				}
				wg.Done()
			}()
		}
		wg.Wait()

		lines := strings.Split(buf.String(), "\n")
		for i := 0; i < groups*linesPerGroup; i++ {
			t.CheckDeepEqual("PREFIX TEXT", lines[i])
		}
	})
}
