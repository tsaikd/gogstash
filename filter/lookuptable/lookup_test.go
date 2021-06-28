package lookuptable

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func writeLookupTableFile(file string, lines []string) error {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fd.Close()

	for _, v := range lines {
		_, err = fd.WriteString(v + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func TestLookupTable(t *testing.T) {
	tmpDir := t.TempDir()
	lookupFile := path.Join(tmpDir, "lookup-test.txt")

	defer t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	lookupLines := []string{
		"192.168.1.1: router",
		"192.168.1.10: httpserver",
		"192.168.1.10\\:443: httpsserver",
	}

	err := writeLookupTableFile(lookupFile, lookupLines)
	if err != nil {
		t.Errorf("write lookup file: %v", err)
		t.Fail()
	}

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(fmt.Sprintf(`
debugch: true
filter:
  - type: lookuptable
    source: hostname
    target: server
    lookup_file: %s
    cache_size: 50
`, lookupFile))))

	if err != nil {
		t.Errorf("parse config: %v", err)
	}

	conf.Start(context.Background())

	tests := []struct {
		name   string
		input  logevent.LogEvent
		want   logevent.LogEvent
		wantOk bool
	}{
		{
			input: logevent.LogEvent{
				Timestamp: time.Time{},
				Message:   "",
				Tags:      nil,
				Extra: map[string]interface{}{
					"hostname": "no-match",
				},
			},
			want: logevent.LogEvent{
				Timestamp: time.Time{},
				Message:   "",
				Tags:      nil,
				Extra: map[string]interface{}{
					"hostname": "no-match",
				},
			},
		},
		{
			input: logevent.LogEvent{
				Timestamp: time.Time{},
				Message:   "",
				Tags:      nil,
				Extra: map[string]interface{}{
					"hostname": "192.168.1.1",
				},
			},
			want: logevent.LogEvent{
				Timestamp: time.Time{},
				Message:   "",
				Tags:      nil,
				Extra: map[string]interface{}{
					"hostname": "192.168.1.1",
					"server":   "router",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.TestInputEvent(tt.input)
			output, err := conf.TestGetOutputEvent(300 * time.Millisecond)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("filter lookuptable got = %v, want %v", output, tt.want)
			}

		})
	}
}

func Test_tokenizeLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [2]string
		wantErr bool
	}{
		{
			name:    "simple key-value, trim",
			input:   "testkey: testvalue \n ",
			want:    [2]string{"testkey", "testvalue"},
			wantErr: false,
		},
		{
			name:    "key-value with escaped semicolons",
			input:   "test.key\\:a\\:b:  testvalue\\:b",
			want:    [2]string{"test.key:a:b", "testvalue:b"},
			wantErr: false,
		},
		{
			name:    "illegal key-value with non-escaped semicolons",
			input:   "testkey:a:testvalue",
			want:    [2]string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizeLine(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenizeLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenizeLine() got = %v, want %v", got, tt.want)
			}
		})
	}
}
