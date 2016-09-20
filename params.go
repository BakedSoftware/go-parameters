// Package parameters parses json into parameters object
// usage:
//   1) parse json to parameters:
// parameters.MakeParsedReq(fn http.HandlerFunc)
//   2) get the parameters:
// params := parameters.GetParams(req)
// val := params.GetXXX("key")
package parameters

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
	"github.com/ugorji/go/codec"
)

const (
	paramsKey = "params"
)

type Params struct {
	isBinary bool
	Values   map[string]interface{}
}

func (p *Params) Get(key string) (interface{}, bool) {
	keys := strings.Split(key, ".")
	root := p.Values
	var ok bool
	var val interface{}
	count := len(keys)
	for i := 0; i < count; i++ {
		val, ok = root[keys[i]]
		if ok && i < count-1 {
			root = val.(map[string]interface{})
		}
	}
	return val, ok
}

func (p *Params) GetFloatOk(key string) (float64, bool) {
	val, ok := p.Get(key)
	if sval, sok := val.(string); sok {
		var err error
		val, err = strconv.ParseFloat(sval, 64)
		ok = err == nil
	}
	if ok {
		return val.(float64), true
	}
	return 0, false
}

func (p *Params) GetFloat(key string) float64 {
	f, _ := p.GetFloatOk(key)
	return f
}

func (p *Params) GetFloatSliceOk(key string) ([]float64, bool) {
	val, ok := p.Get(key)
	if ok {
		switch val.(type) {
		case []float64:
			return val.([]float64), true
		case string:
			raw := strings.Split(val.(string), ",")
			slice := make([]float64, len(raw))
			for i, k := range raw {
				if num, err := strconv.ParseFloat(k, 64); err == nil {
					slice[i] = num
				}
			}
			return slice, true
		case []interface{}:
			raw := val.([]interface{})
			slice := make([]float64, len(raw))
			for i, k := range raw {
				if num, ok := k.(float64); ok {
					slice[i] = num
				} else if num, ok := k.(string); ok {
					if parsed, err := strconv.ParseFloat(num, 64); err == nil {
						slice[i] = parsed
					}
				}
			}
			return slice, true
		}
	}
	return []float64{}, false
}

func (p *Params) GetFloatSlice(key string) []float64 {
	slice, _ := p.GetFloatSliceOk(key)
	return slice
}

func (p *Params) GetBoolOk(key string) (bool, bool) {
	val, ok := p.Get(key)
	if ok {
		if b, ib := val.(bool); ib {
			return b, true
		} else if i, ik := p.GetIntOk(key); ik {
			if i == 0 {
				return false, true
			} else {
				return true, true
			}
		}
	}
	return false, false
}

func (p *Params) GetBool(key string) bool {
	f, _ := p.GetBoolOk(key)
	return f
}

func (p *Params) GetIntOk(key string) (int, bool) {
	val, ok := p.Get(key)
	if sval, sok := val.(string); sok {
		var err error
		val, err = strconv.ParseFloat(sval, 64)
		ok = err == nil
	}
	if ok {
		if ival, ok := val.(int64); ok {
			return int(ival), true
		} else if fval, ok := val.(float64); ok {
			return int(fval), true
		}
	}
	return 0, false
}

func (p *Params) GetInt(key string) int {
	f, _ := p.GetIntOk(key)
	return f
}

func (p *Params) GetInt8Ok(key string) (int8, bool) {
	val, ok := p.GetIntOk(key)

	if !ok || val < math.MinInt8 || val > math.MaxInt8 {
		return 0, false
	}

	return int8(val), true
}

func (p *Params) GetInt8(key string) int8 {
	f, _ := p.GetInt8Ok(key)
	return f
}

func (p *Params) GetInt16Ok(key string) (int16, bool) {
	val, ok := p.GetIntOk(key)

	if !ok || val < math.MinInt16 || val > math.MaxInt16 {
		return 0, false
	}

	return int16(val), true
}

func (p *Params) GetInt16(key string) int16 {
	f, _ := p.GetInt16Ok(key)
	return f
}

func (p *Params) GetInt32Ok(key string) (int32, bool) {
	val, ok := p.GetIntOk(key)

	if !ok || val < math.MinInt32 || val > math.MaxInt32 {
		return 0, false
	}

	return int32(val), true
}

func (p *Params) GetInt32(key string) int32 {
	f, _ := p.GetInt32Ok(key)
	return f
}

func (p *Params) GetInt64Ok(key string) (int64, bool) {
	val, ok := p.GetIntOk(key)

	if !ok {
		return 0, false
	}

	return int64(val), true
}

func (p *Params) GetInt64(key string) int64 {
	f, _ := p.GetIntOk(key)
	return int64(f)
}

func (p *Params) GetIntSliceOk(key string) ([]int, bool) {
	val, ok := p.Get(key)
	if ok {
		switch val.(type) {
		case []int:
			return val.([]int), true
		case []byte:
			val = string(val.([]byte))
			raw := strings.Split(val.(string), ",")
			slice := make([]int, len(raw))
			for i, k := range raw {
				if num, err := strconv.ParseInt(k, 10, 64); err == nil {
					slice[i] = int(num)
				}
			}
			return slice, true
		case string:
			if len(val.(string)) > 0 {
				raw := strings.Split(val.(string), ",")
				slice := make([]int, len(raw))
				for i, k := range raw {
					if num, err := strconv.ParseInt(k, 10, 64); err == nil {
						slice[i] = int(num)
					}
				}
				return slice, true
			}
		case []interface{}:
			raw := val.([]interface{})
			slice := make([]int, len(raw))
			for i, k := range raw {
				if num, ok := k.(int); ok {
					slice[i] = num
				} else if num, ok := k.(float64); ok {
					slice[i] = int(num)
				} else if num, ok := k.(string); ok {
					if parsed, err := strconv.ParseInt(num, 10, 64); err == nil {
						slice[i] = int(parsed)
					}
				}
			}
			return slice, true
		}
	}
	return []int{}, false
}

func (p *Params) GetIntSlice(key string) []int {
	slice, _ := p.GetIntSliceOk(key)
	return slice
}

func (p *Params) GetUint64Ok(key string) (uint64, bool) {
	val, ok := p.Get(key)
	if sval, sok := val.(string); sok {
		var err error
		val, err = strconv.ParseFloat(sval, 64)
		ok = err == nil
	}
	if ok {
		if valInt, ok := val.(int64); ok {
			val = uint64(valInt)
		}
		if valUint, ok := val.(uint64); ok {
			return valUint, true
		} else if valUint, ok := val.(uint); ok {
			return uint64(valUint), true
		} else if valUint, ok := val.(uint8); ok {
			return uint64(valUint), true
		} else if valUint, ok := val.(uint16); ok {
			return uint64(valUint), true
		} else if valUint, ok := val.(uint32); ok {
			return uint64(valUint), true
		} else if valfloat, ok := val.(float64); ok {
			return uint64(valfloat), true
		} else if valbyte, ok := val.([]byte); ok {
			var err error
			val, err = strconv.ParseFloat(string(valbyte), 64)
			ok = err == nil
		}
	}
	return 0, false
}

func (p *Params) GetUint64(key string) uint64 {
	f, _ := p.GetUint64Ok(key)
	return f
}

func (p *Params) GetUint64SliceOk(key string) ([]uint64, bool) {
	if raw, ok := p.GetIntSliceOk(key); ok {
		slice := make([]uint64, len(raw))
		for i, num := range raw {
			slice[i] = uint64(num)
		}
		return slice, true
	}

	return []uint64{}, false
}

func (p *Params) GetUint64Slice(key string) []uint64 {
	slice, _ := p.GetUint64SliceOk(key)
	return slice
}

func (p *Params) GetStringOk(key string) (string, bool) {
	val, ok := p.Get(key)
	if ok {
		if s, is := val.(string); is {
			return s, true
		} else if s, is := val.([]byte); is {
			return string(s), true
		}
	}
	return "", false
}

func (p *Params) GetString(key string) string {
	//Get the string if found
	str, _ := p.GetStringOk(key)

	//Return the string, trim spaces
	return strings.Trim(str, " ")
}

func (p *Params) GetStringSliceOk(key string) ([]string, bool) {
	val, ok := p.Get(key)
	if ok {
		switch val.(type) {
		case []string:
			return val.([]string), true
		case string:
			return strings.Split(val.(string), ","), true
		case []interface{}:
			raw := val.([]interface{})
			slice := make([]string, len(raw))
			for i, k := range raw {
				slice[i] = k.(string)
			}
			return slice, true
		}
	}
	return []string{}, false
}

func (p *Params) GetStringSlice(key string) []string {
	slice, _ := p.GetStringSliceOk(key)
	return slice
}

func (p *Params) GetBytesOk(key string) ([]byte, bool) {
	if dataStr, ok := p.Get(key); ok {
		var dataByte []byte
		var ok bool
		if dataByte, ok = dataStr.([]byte); !ok {
			var err error
			dataByte, err = base64.StdEncoding.DecodeString(dataStr.(string))
			if err != nil {
				log.Println("Error decoding data:", key, err)
				return nil, false
			}
			p.Values[key] = dataByte
		}
		return dataByte, true
	}
	return nil, false
}

func (p *Params) GetBytes(key string) []byte {
	bytes, _ := p.GetBytesOk(key)
	return bytes
}

const (
	DateOnly = "2006-01-02"
	//DateTime is not recommended, rather use time.RFC3339
	DateTime = "2006-01-02 15:04:05"
	// HTMLDateTimeLocal is the format used by the input type datetime-local
	HTMLDateTimeLocal = "2006-01-02T15:04"
)

func (p *Params) GetTimeOk(key string) (time.Time, bool) {
	return p.GetTimeInLocationOk(key, time.UTC)
}

func (p *Params) GetTime(key string) time.Time {
	t, _ := p.GetTimeOk(key)
	return t
}

func (p *Params) GetTimeInLocationOk(key string, loc *time.Location) (time.Time, bool) {
	val, ok := p.Get(key)
	if !ok {
		return time.Time{}, false
	}
	if t, ok := val.(time.Time); ok {
		return t, true
	}
	if str, ok := val.(string); ok {
		if t, err := time.ParseInLocation(time.RFC3339, str, loc); err == nil {
			return t, true
		}
		if t, err := time.ParseInLocation(DateOnly, str, loc); err == nil {
			return t, true
		}
		if t, err := time.ParseInLocation(DateTime, str, loc); err == nil {
			return t, true
		}
		if t, err := time.ParseInLocation(HTMLDateTimeLocal, str, loc); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func (p *Params) GetTimeInLocation(key string, loc *time.Location) time.Time {
	t, _ := p.GetTimeInLocationOk(key, loc)
	return t
}

func (p *Params) GetFileOk(key string) (*multipart.FileHeader, bool) {
	val, ok := p.Get(key)
	if !ok {
		return nil, false
	}
	if fh, ok := val.(*multipart.FileHeader); ok {
		return fh, true
	}
	return nil, false
}

func (p *Params) GetJSONOk(key string) (map[string]interface{}, bool) {
	if v, ok := p.Get(key); ok {
		if d, ok := v.(map[string]interface{}); ok {
			return d, true
		}
	}
	val, ok := p.GetStringOk(key)
	var jsonData map[string]interface{}
	if !ok {
		return jsonData, false
	}
	err := json.NewDecoder(strings.NewReader(val)).Decode(&jsonData)
	if err != nil {
		return jsonData, false
	}
	return jsonData, true
}

func (p *Params) GetJSON(key string) map[string]interface{} {
	data, _ := p.GetJSONOk(key)
	return data
}

func MakeParsedReq(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), paramsKey, ParseParams(r)))
		fn(rw, r)
	}
}

func MakeHTTPRouterParsedReq(fn httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Print("MakeHTTPRouterParsedReq ", r.Header.Get("X-Request-ID"))
		r = r.WithContext(context.WithValue(r.Context(), paramsKey, ParseParams(r)))
		params := GetParams(r)
		for _, param := range p {
			const ID = "id"
			if strings.Contains(param.Key, ID) {
				id, perr := strconv.ParseUint(param.Value, 10, 64)
				if perr != nil {
					params.Values[param.Key] = param.Value
				} else {
					params.Values[param.Key] = id
				}
			} else {
				params.Values[param.Key] = param.Value
			}
		}
		fn(rw, r, p)
	}
}

func GetParams(req *http.Request) *Params {
	params := req.Context().Value(paramsKey).(*Params)
	return params
}

type CustomTypeHandler func(field *reflect.Value, value interface{})

// CustomTypeSetter is used when Imbue is called on an object to handle unknown
// types
var CustomTypeSetter CustomTypeHandler

var (
	typeOfTime      reflect.Type = reflect.TypeOf(time.Time{})
	typeOfPtrToTime reflect.Type = reflect.PtrTo(typeOfTime)
)

//Sets the parameters to the object by type; does not handle nested parameters
func (p *Params) Imbue(obj interface{}) {

	//Get the type of the object
	typeOfObject := reflect.TypeOf(obj).Elem()

	//Get the object
	objectValue := reflect.ValueOf(obj).Elem()

	//Loop our parameters
	for k, _ := range p.Values {

		//Make the incoming key_name into KeyName
		key := SnakeToCamelCase(k, true)

		//Get the type and bool if found
		fieldType, found := typeOfObject.FieldByName(key)

		//Did we get a parameter that is not our object?
		if !found {
			//Error or log
			log.Println("Attempted to set missing param k: ", k, " and key as", key)
			continue
		}

		//Get the field of the key
		field := objectValue.FieldByName(key)

		//Check our types and set accordingly
		if fieldType.Type.Kind() == reflect.String {
			//Set string
			field.Set(reflect.ValueOf(p.GetString(k)))

		} else if fieldType.Type.Kind() == reflect.Uint64 {
			//Set Uint64
			field.Set(reflect.ValueOf(p.GetUint64(k)))

		} else if fieldType.Type.Kind() == reflect.Int {
			//Set Int
			field.Set(reflect.ValueOf(p.GetInt(k)))

		} else if fieldType.Type.Kind() == reflect.Bool {
			//Set bool
			field.Set(reflect.ValueOf(p.GetBool(k)))

		} else if fieldType.Type.Kind() == reflect.Float32 {
			//Set float32
			field.Set(reflect.ValueOf(float32(p.GetFloat(k))))

		} else if fieldType.Type.Kind() == reflect.Float64 {
			//Set float64
			field.Set(reflect.ValueOf(p.GetFloat(k)))
		} else if fieldType.Type == reflect.SliceOf(reflect.TypeOf("")) {
			//Set []string
			field.Set(reflect.ValueOf(p.GetStringSlice(k)))
		} else if fieldType.Type == reflect.SliceOf(reflect.TypeOf(int(0))) {
			//Set []int
			field.Set(reflect.ValueOf(p.GetIntSlice(k)))
		} else if fieldType.Type == reflect.SliceOf(reflect.TypeOf(uint64(0))) {
			//Set []uint64
			field.Set(reflect.ValueOf(p.GetUint64Slice(k)))
		} else if fieldType.Type == reflect.SliceOf(reflect.TypeOf(float64(0))) {
			//Set []float64
			field.Set(reflect.ValueOf(p.GetFloatSlice(k)))
		} else if fieldType.Type == typeOfTime {
			//Set time.Time
			field.Set(reflect.ValueOf(p.GetTime(k)))
		} else if fieldType.Type == typeOfPtrToTime {
			//Set *time.Time
			t := p.GetTime(k)
			field.Set(reflect.ValueOf(&t))
		} else if CustomTypeSetter != nil {
			val, _ := p.Get(k)
			CustomTypeSetter(&field, val)
		}

	}
}

// HasAll will return if all specified keys are found in the params object
func (p *Params) HasAll(keys ...string) (bool, []string) {
	missing := make([]string, 0)
	for _, key := range keys {
		if _, exists := p.Values[key]; !exists {
			missing = append(missing, key)
		}
	}
	return len(missing) == 0, missing
}

//Permits only the allowed fields given by allowedKeys
func (p *Params) Permit(allowedKeys []string) {
	for key, _ := range p.Values {
		if !contains(allowedKeys, key) {
			delete(p.Values, key)
		}
	}
}

func contains(haystack []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, straw := range haystack {
		if strings.ToLower(straw) == needle {
			return true
		}
	}
	return false
}

func ParseParams(req *http.Request) *Params {
	var p Params
	ct := req.Header.Get("Content-Type")
	ct = strings.Split(ct, ";")[0]
	if ct == "multipart/form-data" {
		if err := req.ParseMultipartForm(10000000); err != nil {
			log.Println("Request.ParseMultipartForm Error", err)
		}
	} else {
		if err := req.ParseForm(); err != nil {
			log.Println("Request.ParseForm Error", err)
		}
	}
	tmap := make(map[string]interface{}, len(req.Form))
	for k, v := range req.Form {
		if strings.ToLower(v[0]) == "true" {
			tmap[k] = true
		} else if strings.ToLower(v[0]) == "false" {
			tmap[k] = false
		} else {
			tmap[k] = v[0]
		}
	}

	if req.MultipartForm != nil {
		for k, v := range req.MultipartForm.File {
			tmap[k] = v[0]
		}
	}

	if ct == "application/json" && req.ContentLength > 0 {
		err := json.NewDecoder(req.Body).Decode(&p.Values)
		if err != nil {
			log.Println("Content-Type is \"application/json\" but no valid json data received", err)
			p.Values = tmap
		}
		for k, v := range tmap {
			if _, pres := p.Values[k]; !pres {
				p.Values[k] = v
			}
		}
	} else if ct == "application/x-msgpack" {
		var mh codec.MsgpackHandle
		p.isBinary = true
		mh.MapType = reflect.TypeOf(p.Values)
		body, _ := ioutil.ReadAll(req.Body)
		if len(body) > 0 {
			buff := bytes.NewBuffer(body)
			first := body[0]
			if (first >= 0x80 && first <= 0x8f) || (first == 0xde || first == 0xdf) {
				err := codec.NewDecoder(buff, &mh).Decode(&p.Values)
				if err != nil && err != io.EOF {
					log.Println("Failed decoding msgpack", err)
				}
			} else {
				if p.Values == nil {
					p.Values = make(map[string]interface{}, 0)
				}
				var err error
				for err == nil {
					vals := make([]interface{}, 0)
					err = codec.NewDecoder(buff, &mh).Decode(&vals)
					if err != nil && err != io.EOF {
						log.Println("Failed decoding msgpack", err)
					} else {
						for i := len(vals) - 1; i >= 1; i -= 2 {
							p.Values[string(vals[i-1].([]byte))] = vals[i]
						}
					}
				}
			}
		} else {
			p.Values = make(map[string]interface{}, 0)
		}
		for k, v := range tmap {
			if _, pres := p.Values[k]; !pres {
				p.Values[k] = v
			}
		}
	} else {
		p.Values = tmap
	}

	for k, v := range mux.Vars(req) {
		const ID = "id"
		if strings.Contains(k, ID) {
			id, perr := strconv.ParseUint(v, 10, 64)
			if perr != nil {
				p.Values[k] = v
			} else {
				p.Values[k] = id
			}
		} else {
			p.Values[k] = v
		}
	}

	return &p
}
