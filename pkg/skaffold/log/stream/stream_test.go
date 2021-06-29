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
		var buf bytes.Buffer

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				for i := 0; i < 100; i++ {
					// printLogLine(output.Default, &buf, func() bool { return false }, &sync.Mutex{}, "", output.Default.Sprintf("%s ", "PREFIX")+"TEXT\n")
					printLogLine(output.Default, &buf, func() bool { return false }, &sync.Mutex{}, "PREFIX", "TEXT\n")
				}
				wg.Done()
			}()
		}
		wg.Wait()

		lines := strings.Split(buf.String(), "\n")
		for i := 0; i < 5*100; i++ {
			t.CheckDeepEqual("PREFIX TEXT", lines[i])
		}
	})
}
