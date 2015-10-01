package parameters

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type Params struct {
	Values map[string]interface{}
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

func (p *Params) GetBoolOk(key string) (bool, bool) {
	val, ok := p.Get(key)
	if ok {
		if b, ib := val.(bool); ib {
			return b, true
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

func (p *Params) GetUint64Ok(key string) (uint64, bool) {
	val, ok := p.Get(key)
	if sval, sok := val.(string); sok {
		var err error
		val, err = strconv.ParseFloat(sval, 64)
		ok = err == nil
	}
	if ok {
		if valUint, ok := val.(uint64); ok {
			return valUint, true
		} else if valfloat, ok := val.(float64); ok {
			return uint64(valfloat), true
		}
	}
	return 0, false
}

func (p *Params) GetUint64(key string) uint64 {
	f, _ := p.GetUint64Ok(key)
	return f
}

func (p *Params) GetStringOk(key string) (string, bool) {
	val, ok := p.Get(key)
	if ok {
		if s, is := val.(string); is {
			return s, true
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

func (p *Params) GetTimeOk(key string) (time.Time, bool) {
	val, ok := p.Get(key)
	if !ok {
		return time.Time{}, false
	}
	if t, ok := val.(time.Time); ok {
		return t, true
	}
	if str, ok := val.(string); ok {
		if t, err := time.Parse("2006-01-02", str); err == nil {
			return t, true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func (p *Params) GetTime(key string) time.Time {
	t, _ := p.GetTimeOk(key)
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
	return func(rw http.ResponseWriter, req *http.Request) {
		ParseParams(req)
		fn(rw, req)
	}
}

func GetParams(req *http.Request) *Params {
	params := context.Get(req, "params").(Params)
	return &params
}

type CustomTypeHandler func(field *reflect.Value, value interface{})

// CustomTypeSetter is used when Imbue is called on an object to handle unknown
// types
var CustomTypeSetter CustomTypeHandler

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

		} else if CustomTypeSetter != nil {
			val, _ := p.Get(k)
			CustomTypeSetter(&field, val)
		}

	}
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

func ParseParams(req *http.Request) {
	var p Params
	req.ParseMultipartForm(10000000)
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

	ct := req.Header.Get("Content-Type")
	ct = strings.Split(ct, ";")[0]
	if ct == "application/json" && req.ContentLength > 0 {
		err := json.NewDecoder(req.Body).Decode(&p.Values)
		if err != nil {
			log.Println("Decode:", err)
			p.Values = tmap
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
		}
	}

	context.Set(req, "params", p)
}
