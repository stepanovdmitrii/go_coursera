package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type XmlUser struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func (XmlUser *XmlUser) ToUser() *User {
	result := &User{}
	result.Id = XmlUser.Id
	result.Name = XmlUser.FirstName + XmlUser.LastName
	result.Age = XmlUser.Age
	result.About = XmlUser.About
	result.Gender = XmlUser.Gender
	return result
}

type XmlData struct {
	XMLName xml.Name  `xml:"root"`
	Users   []XmlUser `xml:"row"`
}

var initialized bool = false
var useLimit bool = true
var users []User = nil

func InitUsers() {
	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		panic("failed to open dataset")
	}

	defer xmlFile.Close()

	bytes, err := ioutil.ReadAll(xmlFile)

	if err != nil {
		panic("failed to read dataset")
	}
	var xmlData XmlData
	xml.Unmarshal(bytes, &xmlData)
	for _, u := range xmlData.Users {
		users = append(users, *(u.ToUser()))
	}
}

func ParseSearchRequest(r *http.Request) (*SearchRequest, error) {
	query := r.URL.Query()
	result := &SearchRequest{}

	lim := query.Get("limit")
	limValue, _ := strconv.Atoi(lim)
	result.Limit = limValue

	off := query.Get("offset")
	offValue, _ := strconv.Atoi(off)
	result.Offset = offValue

	result.Query = query.Get("query")

	orderField := query.Get("order_field")
	result.OrderField = orderField

	orderBy := query.Get("order_by")
	orderByValue, _ := strconv.Atoi(orderBy)
	result.OrderBy = orderByValue

	return result, nil
}

func Match(user *User, request *SearchRequest) bool {
	if request.Query == "" {
		return true
	}
	return strings.Contains(user.Name, request.Query) || strings.Contains(user.About, request.Query)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if false == initialized {
		InitUsers()
		initialized = true
	}

	request, _ := ParseSearchRequest(r)

	if request.OrderField == "" {
		request.OrderField = "Name"
	}

	if request.OrderField != "Id" && request.OrderField != "Name" && request.OrderField != "Age" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
		return
	}

	skip := request.Offset
	result := []User{}

	for _, u := range users {

		if useLimit && len(result) == request.Limit {
			break
		}

		if Match(&u, request) {
			if skip > 0 {
				skip--
				continue
			}

			result = append(result, u)
		}
	}

	if request.OrderField == "Id" {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Id < result[j].Id
		})
	}

	if request.OrderField == "Name" {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})
	}

	if request.OrderField == "Age" {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Age < result[j].Age
		})
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func UnknownError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"Error": "Unknown error"}`)
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	io.WriteString(w, `{"Error": "Bad token"}`)
}

func FatalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, `{"Error": "StatusInternalServerError"}`)
}

func BadOrderField(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
}

func BadFormat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `not json`)
}

func OkButBadFormat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `not json`)
}

func Timeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
}

func Stub(w http.ResponseWriter, r *http.Request) {

}

func Test_Limit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Limit = -1
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("limit lower than zero must not be accepted")
	}

	req.Limit = 25
	resp, err := client.FindUsers(*req)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(resp.Users) != 25 {
		t.Errorf("invalid users count for limit = 25")
		return
	}

	req.Limit = 30
	resp, err = client.FindUsers(*req)
	if len(resp.Users) != 25 {
		t.Errorf("invalid users count for limit = 30")
		return
	}

	useLimit = false
	resp, err = client.FindUsers(*req)
	if len(resp.Users) != 35 {
		t.Errorf("invalid users count for limit = 30 and useLimit = false %v", len(resp.Users))
		return
	}
	useLimit = true
}

func Test_Offset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = -1
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("offset lower than zero must not be accepted")
	}
}

func Test_UnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(UnknownError))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Unauthorized))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_FatalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FatalError))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_BadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(BadOrderField))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_BadFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(BadFormat))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_OkButBadFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(OkButBadFormat))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_MoreThanLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(OkButBadFormat))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Timeout))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ts.URL
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}

func Test_NilUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Stub))
	defer ts.Close()
	client := &SearchClient{}
	client.URL = ""
	req := &SearchRequest{}

	req.Offset = 20
	_, err := client.FindUsers(*req)

	if err == nil {
		t.Errorf("shoul pass error")
	}
}
