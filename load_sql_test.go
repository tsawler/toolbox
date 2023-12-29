package toolbox

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

var tools Tools

func TestLoadSQLQueries(t *testing.T) {
	tests := []struct {
		fileName string
		key      string
		value    string
		equal    bool
		err      bool
	}{
		{fileName: "./testdata/not.sql", key: "", value: "", equal: true, err: true},
		{fileName: "./testdata/not.sql", key: "not", value: "", equal: true, err: true},
		{fileName: "./testdata/not.sql", key: "not", value: "equal", equal: false, err: true},
		{fileName: "./testdata/test.sql", key: "TEST1", value: "WHERE ass.id=$1;", equal: true, err: false},
		{fileName: "./testdata/test.sql", key: "TEST1", value: "WHERE ass.id=$1", equal: false, err: false},
		{fileName: "./testdata/test.sql", key: "TEST2", value: "id = $1;", equal: true, err: false},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			query, err := tools.LoadSQLQueries(tt.fileName)
			if (err != nil) != tt.err {
				t.Errorf("LoadSQLQueries() error: %v, except: %v", err, tt.err)
			}
			if strings.HasSuffix(query[tt.key], tt.value) != tt.equal {
				t.Errorf("LoadSQLQueries() error: %v, except: %v", err, tt.equal)
			}
		})
	}
}

func TestParseSQLQueries(t *testing.T) {
	file, err := os.Open("./testdata/test.sql")
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	if (err != nil) != false {
		t.Errorf("File Open result: %v, expect: %v", false, true)
	}
	_, err = parseSQLQueries(file, make(map[string]string))
	if (err != nil) != false {
		t.Errorf("parseSQLQueries() result: %v, expect: %v", false, true)
	}
}

func TestIsSQLQuery(t *testing.T) {
	if result := isSQLQuery("-- "); result != true {
		t.Errorf("isSQLQuery() result: %v, expect: %v", result, true)
	}
	if result := isSQLQuery("--"); result != false {
		t.Errorf("isSQLQuery() result: %v, expect: %v", result, false)
	}
}

func TestExtractKey(t *testing.T) {
	tests := []struct {
		value  string
		expect string
	}{
		{value: "-- ABC", expect: "ABC"},
		{value: "DEF", expect: ""},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if result := extractKey(tt.value); result != tt.expect {
				t.Errorf("extractKey() result: %v, expect: %v", result, tt.expect)
			}
		})
	}
}

func TestHasPrefixInList(t *testing.T) {
	type args struct {
		key   string
		value []string
	}
	tests := []struct {
		args   args
		expect bool
	}{
		{args: args{key: "abc"}, expect: false},
		{args: args{key: "abc", value: []string{"abc", "def"}}, expect: true},
		{args: args{key: "xyz", value: []string{"abc", "def"}}, expect: false},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if result := hasPrefixInList(tt.args.key, tt.args.value); result != tt.expect {
				t.Errorf("hasPrefixInList() result: %v, expect: %v", result, tt.expect)
			}
		})
	}
}
