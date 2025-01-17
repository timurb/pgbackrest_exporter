package backrest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/common/promlog"
)

var (
	mockStdout, mockStderr string
	mockExit               int
	logger                 = getLogger()
)

func TestSetPromPortandPath(t *testing.T) {
	var (
		testPort          = "9854"
		testEndpoit       = "/metrics"
		testTLSConfigPath = ""
	)
	SetPromPortandPath(testPort, testEndpoit, testTLSConfigPath)
	if testPort != promPort || testEndpoit != promEndpoint || testTLSConfigPath != promTLSConfigPath {
		t.Errorf("\nVariables do not match: %s,\nwant: %s;\nendpoint: %s,\nwant: %s", testPort, promPort, testEndpoit, promEndpoint)
	}
}

func TestGetPgBackRestInfo(t *testing.T) {
	type args struct {
		config            string
		configIncludePath string
		stanzas           []string
		stanzasExclude    []string
		verbose           bool
	}
	tests := []struct {
		name       string
		args       args
		mockStdout string
		mockStderr string
		mockExit   int
		testText   string
	}{
		{"GetPgBackRestInfoGoodDataReturn",
			args{"", "", []string{""}, []string{""}, false},
			`[{"archive":[{"database":{"id":1,"repo-key":1},"id":"13-1",` +
				`"max":"000000010000000000000002","min":"000000010000000000000001"}],` +
				`"backup":[{"archive":{"start":"000000010000000000000002","stop":"000000010000000000000002"},` +
				`"backrest":{"format":5,"version":"2.34"},"database":{"id":1,"repo-key":1},` +
				`"info":{"delta":24316343,"repository":{"delta":2969512,"size":2969512},"size":24316343},` +
				`"label":"20210614-213200F","prior":null,"reference":null,"timestamp":{"start":1623706320,` +
				`"stop":1623706322},"type":"full"}],"cipher":"none","db":[{"id":1,"repo-key":1,` +
				`"system-id":6970977677138971135,"version":"13"}],"name":"demo","repo":[{"cipher":"none",` +
				`"key":1,"status":{"code":0,"message":"ok"}}],"status":{"code":0,"lock":{"backup":` +
				`{"held":false}},"message":"ok"}}]`,
			``,
			0,
			""},
		{"GetPgBackRestInfoGoodDataReturnWithWarn",
			args{"", "", []string{""}, []string{""}, false},
			`[{"archive":[{"database":{"id":1,"repo-key":1},"id":"13-1",` +
				`"max":"000000010000000000000002","min":"000000010000000000000001"}],` +
				`"backup":[{"archive":{"start":"000000010000000000000002","stop":"000000010000000000000002"},` +
				`"backrest":{"format":5,"version":"2.34"},"database":{"id":1,"repo-key":1},` +
				`"info":{"delta":24316343,"repository":{"delta":2969512,"size":2969512},"size":24316343},` +
				`"label":"20210614-213200F","prior":null,"reference":null,"timestamp":{"start":1623706320,` +
				`"stop":1623706322},"type":"full"}],"cipher":"none","db":[{"id":1,"repo-key":1,` +
				`"system-id":6970977677138971135,"version":"13"}],"name":"demo","repo":[{"cipher":"none",` +
				`"key":1,"status":{"code":0,"message":"ok"}}],"status":{"code":0,"lock":{"backup":` +
				`{"held":false}},"message":"ok"}}]`,
			`WARN: environment contains invalid option 'test'`,
			0,
			`msg="pgBackRest message" err="WARN: environment contains invalid option 'test'`},
		{"GetPgBackRestInfoBadDataReturn",
			args{"", "", []string{""}, []string{""}, false},
			``,
			`msg="pgBackRest message" err="ERROR: [029]: missing '=' in key/value at line 9: test"`,
			29,
			`msg="Get data from pgBackRest failed" err="exit status 29`},
		{"GetPgBackRestInfoZeroDataReturn",
			args{"", "", []string{""}, []string{""}, false},
			`[]`,
			``,
			0,
			`msg="No backup data returned"`},
		{"GetPgBackRestInfoJsonUnmarshalFail",
			args{"", "", []string{""}, []string{""}, false},
			`[{}`,
			``,
			0,
			`msg="Parse JSON failed" err="unexpected end of JSON input"`},
		{"GetPgBackRestInfoEqualIncludeExcludeLists",
			args{"", "", []string{"demo"}, []string{"demo"}, false},
			``,
			``,
			0,
			`msg="Stanza is specified in include and exclude lists" stanza=demo`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetMetrics()
			mockStdout = tt.mockStdout
			mockStderr = tt.mockStderr
			execCommand = fakeExecCommand
			mockExit = tt.mockExit
			defer func() { execCommand = exec.Command }()
			out := &bytes.Buffer{}
			lc := log.NewLogfmtLogger(out)
			GetPgBackRestInfo(
				tt.args.config,
				tt.args.configIncludePath,
				tt.args.stanzas,
				tt.args.stanzasExclude,
				tt.args.verbose,
				lc,
			)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockExit)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockStdout,
		"STDERR=" + mockStderr,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "%s", os.Getenv("STDOUT"))
	fmt.Fprintf(os.Stderr, "%s", os.Getenv("STDERR"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

// Set logger for tests.
// If it's necessary to capture the logs output in the test,
// a separate logger is used inside the test.
// The info logging level is used.
func getLogger() log.Logger {
	var err error
	logLevel := &promlog.AllowedLevel{}
	err = logLevel.Set("info")
	if err != nil {
		panic(err)
	}
	promlogConfig := &promlog.Config{}
	promlogConfig.Level = logLevel
	if err != nil {
		panic(err)
	}
	return promlog.New(promlogConfig)
}
