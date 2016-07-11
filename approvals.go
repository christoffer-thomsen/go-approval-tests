package ApprovalTests_go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/approvals/ApprovalTests_go/reporters"
)

var (
	defaultReporter            *reporters.Reporter = reporters.NewDiffReporter()
	defaultFrontLoadedReporter *reporters.Reporter = reporters.NewFrontLoadedReporter()
)

type Failable interface {
	Fail()
}

func VerifyWithExtension(t Failable, reader io.Reader, extWithDot string) error {
	namer, err := getApprovalName()
	if err != nil {
		return err
	}

	reporter := getReporter()
	err = namer.compare(namer.getApprovalFile(extWithDot), namer.getReceivedFile(extWithDot), reader)
	if err != nil {
		reporter.Report(namer.getApprovalFile(extWithDot), namer.getReceivedFile(extWithDot))
		t.Fail()
	} else {
		os.Remove(namer.getReceivedFile(extWithDot))
	}

	return err
}

func Verify(t Failable, reader io.Reader) error {
	return VerifyWithExtension(t, reader, ".txt")
}

func VerifyString(t Failable, s string) {
	reader := strings.NewReader(s)
	Verify(t, reader)
}

func VerifyJSONBytes(t Failable, bs []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(bs, &obj)
	if err != nil {
		message := fmt.Sprintf("error while parsing JSON\nerror:\n  %s\nJSON:\n  %s\n", err, string(bs))
		return VerifyWithExtension(t, strings.NewReader(message), ".json")
	}
	jsonb, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		message := fmt.Sprintf("error while pretty printing JSON\nerror:\n  %s\nJSON:\n  %s\n", err, string(bs))
		return VerifyWithExtension(t, strings.NewReader(message), ".json")
	}

	return VerifyWithExtension(t, bytes.NewReader(jsonb), ".json")
}

type reporterCloser struct {
	reporter *reporters.Reporter
}

func (s *reporterCloser) Close() error {
	defaultReporter = s.reporter
	return nil
}

type frontLoadedReporterCloser struct {
	reporter *reporters.Reporter
}

func (s *frontLoadedReporterCloser) Close() error {
	defaultFrontLoadedReporter = s.reporter
	return nil
}

// Add at the test or method level to configure your reporter.
//
// The following examples shows how to use a reporter for all of your test cases
// through go's setup feature.
//
// func TestMain(m *testing.M) {
// 	r := UseReporter(reporters.NewBeyondCompareReporter())
//      defer r.Close()
//
// 	m.Run()
// }
//
func UseReporter(reporter reporters.Reporter) io.Closer {
	closer := &reporterCloser{
		reporter: defaultReporter,
	}

	defaultReporter = &reporter
	return closer
}

func UseFrontLoadedReporter(reporter reporters.Reporter) io.Closer {
	closer := &frontLoadedReporterCloser{
		reporter: defaultFrontLoadedReporter,
	}

	defaultFrontLoadedReporter = &reporter
	return closer
}

func getReporter() reporters.Reporter {
	return reporters.NewFirstWorkingReporter(
		*defaultFrontLoadedReporter,
		*defaultReporter,
	)
}
