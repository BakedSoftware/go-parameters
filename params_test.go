package parameters

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestParseJSONBody(t *testing.T) {
	body := "{ \"test\": true }"
	r, err := http.NewRequest("POST", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/json")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}
}

func TestParseJSONBodyContentType(t *testing.T) {
	body := "{ \"test\": true }"
	r, err := http.NewRequest("POST", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/json; charset=utf8")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}
}

func TestParseNestedJSONBody(t *testing.T) {
	body := "{ \"test\": true, \"coord\": { \"lat\": 50.505, \"lon\": 10.101 }}"
	r, err := http.NewRequest("POST", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/json")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}

	val, present = params.Get("coord")
	if !present {
		t.Fatal("Key: 'coord' not found")
	}

	coord := val.(map[string]interface{})

	lat, present := coord["lat"]
	if !present {
		t.Fatal("Key: 'lat' not found")
	}
	if lat != 50.505 {
		t.Fatal("Value of 'lat' should be 50.505, got: ", lat)
	}

	lat, present = params.Get("coord.lat")
	if !present {
		t.Fatal("Nested Key: 'lat' not found")
	}
	if lat != 50.505 {
		t.Fatal("Value of 'lat' should be 50.505, got: ", lat)
	}

	lon, present := coord["lon"]
	if !present {
		t.Fatal("Key: 'lon' not found")
	}
	if lon != 10.101 {
		t.Fatal("Value of 'lon' should be 10.101, got: ", lon)
	}

	lon, present = params.Get("coord.lon")
	if !present {
		t.Fatal("Nested Key: 'lon' not found")
	}
	if lon != 10.101 {
		t.Fatal("Value of 'lon' should be 10.101, got: ", lon)
	}
}

func TestParseGET(t *testing.T) {
	body := ""
	r, err := http.NewRequest("GET", "test?test=true", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}
}

func TestParsePOST(t *testing.T) {
	body := "test=true"
	r, err := http.NewRequest("POST", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}
}

func TestParsePUT(t *testing.T) {
	body := "test=true"
	r, err := http.NewRequest("PUT", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}
}

func TestParsePostUrlJSON(t *testing.T) {
	body := "{\"test\":true}"
	r, err := http.NewRequest("PUT", "test?test=false&id=1", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/json")

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}

	val, present = params.GetFloatOk("id")
	if !present {
		t.Fatal("Key: 'id' not found")
	}
	if val != 1.0 {
		t.Fatal("Value of 'id' should be 1, got: ", val)
	}
}

func TestParseJSONBodyMux(t *testing.T) {
	body := "{ \"test\": true }"
	r, err := http.NewRequest("POST", "/test/42", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/json")
	m := mux.NewRouter()
	m.KeepContext = true
	m.HandleFunc("/test/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
	})

	var match mux.RouteMatch
	if !m.Match(r, &match) {
		t.Error("Mux did not match")
	}
	m.ServeHTTP(nil, r)

	ParseParams(r)

	params := GetParams(r)

	val, present := params.Get("test")
	if !present {
		t.Fatal("Key: 'test' not found")
	}
	if val != true {
		t.Fatal("Value of 'test' should be 'true', got: ", val)
	}

	val, present = params.Get("id")
	if !present {
		t.Fatal("Key: 'id' not found")
	}
	if val != uint64(42) {
		t.Fatal("Value of 'id' should be 42, got: ", val)
	}
}

func TestImbue(t *testing.T) {
	body := "test=true&keys=this,that,something&values=1,2,3"
	r, err := http.NewRequest("PUT", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ParseParams(r)

	params := GetParams(r)

	type testType struct {
		Test   bool
		Keys   []string
		Values []int
	}

	var obj testType
	params.Imbue(&obj)

	if obj.Test != true {
		t.Fatal("Value of 'test' should be 'true', got: ", obj.Test)
	}
	if len(obj.Keys) != 3 {
		t.Fatal("Length of 'keys' should be '3', got: ", len(obj.Keys))
	}
	if len(obj.Values) != 3 {
		t.Fatal("Length of 'values' should be '3', got: ", len(obj.Values))
	}
	values := []int{1, 2, 3}
	for i, k := range obj.Values {
		if values[i] != k {
			t.Log("Expected ", values[i], ", got:", k)
			t.Fail()
		}
	}
}

func TestImbueTime(t *testing.T) {
	body := "test=true&created_at=2016-06-07T00:30Z&remind_on=2016-07-17"
	r, err := http.NewRequest("PUT", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ParseParams(r)

	params := GetParams(r)

	type testType struct {
		Test      bool
		CreatedAt time.Time
		RemindOn  *time.Time
	}

	var obj testType
	params.Imbue(&obj)

	if obj.Test != true {
		t.Fatal("Value of 'test' should be 'true', got: ", obj.Test)
	}
	createdAt, _ := time.Parse(time.RFC3339, "2016-06-07T00:30Z00:00")
	if !obj.CreatedAt.Equal(createdAt) {
		t.Fatal("CreatedAt should be '2016-06-07T00:30Z', got:", obj.CreatedAt)
	}
	remindOn, _ := time.Parse(DateOnly, "2016-07-17")
	if obj.RemindOn == nil || !obj.RemindOn.Equal(remindOn) {
		t.Fatal("RemindOn should be '2016-07-17', got:", obj.RemindOn)
	}
}

func TestHasAll(t *testing.T) {
	body := "test=true&keys=this,that,something&values=1,2,3"
	r, err := http.NewRequest("PUT", "test", strings.NewReader(body))
	if err != nil {
		t.Fatal("Could not build request", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ParseParams(r)

	params := GetParams(r)
	//Test All
	if ok, missing := params.HasAll("test", "keys", "values"); !ok || len(missing) > 0 {
		t.Fatal("Params should have all keys, could not find", missing)
	}

	// Test Partial Contains
	if ok, missing := params.HasAll("test"); !ok || len(missing) > 0 {
		t.Fatal("Params should have key 'test', could not find", missing)
	}

	// Test Partial Missing
	if ok, missing := params.HasAll("test", "nope"); ok || len(missing) == 0 {
		t.Fatal("Params should not have key 'nope'", missing)
	} else if contains(missing, "test") {
		t.Fatal("Missing should not contain 'test'")
	}

	// Test All missing
	if ok, missing := params.HasAll("negative", "nope"); ok || len(missing) == 0 {
		t.Fatal("Params should not have key 'nope' nor 'negative'", missing)
	}
}
