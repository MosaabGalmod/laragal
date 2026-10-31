package winutil

import (
	"reflect"
	"testing"
)

func TestParseListeningPIDs(t *testing.T) {
	sampleOutput := `
Active Connections

  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1001
  TCP    0.0.0.0:9085           0.0.0.0:0              LISTENING       2001
  TCP    0.0.0.0:90850          0.0.0.0:0              LISTENING       9999
  TCP    127.0.0.1:9085         0.0.0.0:0              LISTENING       2001
  TCP    127.0.0.1:19085        0.0.0.0:0              LISTENING       8888
  TCP    [::]:9085              [::]:0                 LISTENING       2002
  TCP    127.0.0.1:9085         127.0.0.1:54321        ESTABLISHED     2001
  TCP    0.0.0.0:80             0.0.0.0:0              LISTENING       4
  UDP    0.0.0.0:9085           *:*                                    5555
`

	// 1. Exact port match for 9085 should find PIDs 2001 and 2002, deduplicated, ignoring 90850, 19085, and ESTABLISHED
	got := ParseListeningPIDs(sampleOutput, 9085)
	expected := []int{2001, 2002}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ParseListeningPIDs(9085) = %v, expected %v", got, expected)
	}

	// 2. Exact match for 90850 should only find 9999
	got90850 := ParseListeningPIDs(sampleOutput, 90850)
	expected90850 := []int{9999}
	if !reflect.DeepEqual(got90850, expected90850) {
		t.Errorf("ParseListeningPIDs(90850) = %v, expected %v", got90850, expected90850)
	}

	// 3. Port with no listeners should return nil/empty
	gotNone := ParseListeningPIDs(sampleOutput, 7777)
	if len(gotNone) != 0 {
		t.Errorf("ParseListeningPIDs(7777) = %v, expected empty", gotNone)
	}

	// 4. Invalid port <= 0
	gotZero := ParseListeningPIDs(sampleOutput, 0)
	if len(gotZero) != 0 {
		t.Errorf("ParseListeningPIDs(0) = %v, expected empty", gotZero)
	}
}
