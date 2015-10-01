package parameters

import "testing"

func TestCamelCaseToSnakeCase(t *testing.T) {
	entries := map[string]string{
		"ID":          "id",
		"User":        "user",
		"UserName":    "user_name",
		"UserID":      "user_id",
		"MyJSON":      "my_json",
		"ProfileHTML": "profile_html",
		"RequestXML":  "request_xml",
	}

	for k, v := range entries {
		transformed := CamelToSnakeCase(k)
		if transformed != v {
			t.Logf(`Expected "%s" to become "%s", not "%s"`, k, v, transformed)
			t.Fail()
		}
	}
}

func TestSnakeCaseToCamelCase(t *testing.T) {
	entries := map[string]string{
		"id":           "ID",
		"user":         "User",
		"user_name":    "UserName",
		"user_id":      "UserID",
		"my_json":      "MyJSON",
		"profile_html": "ProfileHTML",
		"request_xml":  "RequestXML",
	}

	for k, v := range entries {
		transformed := SnakeToCamelCase(k, true)
		if transformed != v {
			t.Logf(`Expected "%s" to become "%s", not "%s"`, k, v, transformed)
			t.Fail()
		}
	}
}
