package main

import "fmt"
import "strconv"
import "strings"
import "encoding/json"
import "net/http"

func (in *ProfileParams) ApiGenSet(r *http.Request) error {

	// Login


	LoginRaw := r.URL.Query().Get("login")
	if r.Method == "POST" {
		LoginRaw = r.FormValue("login")
	}

	if true && LoginRaw == "" {
		return fmt.Errorf("login must me not empty")
	}

	if false && LoginRaw == "" {
		LoginRaw = ""
	}

	if false {
		isValid := false
		for _, enumValue := range strings.Split("", "|") {
			isValid = isValid || LoginRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("login must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if false && len(LoginRaw) < 0 {
		return fmt.Errorf("login len must be >= 0")
	}
	in.Login = LoginRaw
    return nil
}

func (in *CreateParams) ApiGenSet(r *http.Request) error {

	// Login


	LoginRaw := r.URL.Query().Get("login")
	if r.Method == "POST" {
		LoginRaw = r.FormValue("login")
	}

	if true && LoginRaw == "" {
		return fmt.Errorf("login must me not empty")
	}

	if false && LoginRaw == "" {
		LoginRaw = ""
	}

	if false {
		isValid := false
		for _, enumValue := range strings.Split("", "|") {
			isValid = isValid || LoginRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("login must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if true && len(LoginRaw) < 10 {
		return fmt.Errorf("login len must be >= 10")
	}
	in.Login = LoginRaw

	// Name


	NameRaw := r.URL.Query().Get("full_name")
	if r.Method == "POST" {
		NameRaw = r.FormValue("full_name")
	}

	if false && NameRaw == "" {
		return fmt.Errorf("full_name must me not empty")
	}

	if false && NameRaw == "" {
		NameRaw = ""
	}

	if false {
		isValid := false
		for _, enumValue := range strings.Split("", "|") {
			isValid = isValid || NameRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("full_name must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if false && len(NameRaw) < 0 {
		return fmt.Errorf("full_name len must be >= 0")
	}
	in.Name = NameRaw

	// Status


	StatusRaw := r.URL.Query().Get("status")
	if r.Method == "POST" {
		StatusRaw = r.FormValue("status")
	}

	if false && StatusRaw == "" {
		return fmt.Errorf("status must me not empty")
	}

	if true && StatusRaw == "" {
		StatusRaw = "user"
	}

	if true {
		isValid := false
		for _, enumValue := range strings.Split("user|moderator|admin", "|") {
			isValid = isValid || StatusRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("status must be one of [" + strings.Join(strings.Split("user|moderator|admin", "|"), ", ") + "]")
		}
	}

	if false && len(StatusRaw) < 0 {
		return fmt.Errorf("status len must be >= 0")
	}
	in.Status = StatusRaw

	// Age
	AgeRaw := r.URL.Query().Get("age")
	if r.Method == "POST" {
		AgeRaw = r.FormValue("age")
	}

	var AgeValue int

	if false && AgeRaw == "" {
		return fmt.Errorf("age must me not empty")
	}

	if AgeRaw != "" {
		AgeConvValue, AgeErr := strconv.Atoi(AgeRaw)
		if AgeErr !=nil {
			return fmt.Errorf("age must be int")
		}
		AgeValue = AgeConvValue
	}

	if false && AgeRaw == "" {
		if "" != "" {
			AgeValue, _ = strconv.Atoi("")
		}
		
	}

	if false {
		isValid := false

		for _, enumValue := range strings.Split("", "|") {
			enumIntValue, _ := strconv.Atoi(enumValue)
			isValid = isValid || AgeValue == enumIntValue
		}

		if !isValid {
			return fmt.Errorf("age must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if true && AgeValue < 0 {
		return fmt.Errorf("age must be >= 0")
	}

	if true && AgeValue > 128 {
		return fmt.Errorf("age must be <= 128")
	}

	in.Age = AgeValue
    return nil
}


func (api *MyApi) wrapperApiGenProfile(w http.ResponseWriter, r *http.Request){
	if "" != "" && r.Method != "" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\" : \"bad method\"}"))
		return
	}

	if false && r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\" : \"unauthorized\"}"))
		return
	}

	var p ProfileParams
	setErr := p.ApiGenSet(r)
	if setErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + strings.ToLower(setErr.Error()) +"\"}"))
		return
	}
	
	res, err := api.Profile(r.Context(), p)

	if err != nil {
		switch err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(apiErr.Err.Error()) +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(err.Error()) +"\"}"))
		}
		return
	}

	mapResponse := map[string]interface{}{"error": "", "response": res}

	data, _ := json.Marshal(mapResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (api *MyApi) wrapperApiGenCreate(w http.ResponseWriter, r *http.Request){
	if "POST" != "" && r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\" : \"bad method\"}"))
		return
	}

	if true && r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\" : \"unauthorized\"}"))
		return
	}

	var p CreateParams
	setErr := p.ApiGenSet(r)
	if setErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + strings.ToLower(setErr.Error()) +"\"}"))
		return
	}
	
	res, err := api.Create(r.Context(), p)

	if err != nil {
		switch err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(apiErr.Err.Error()) +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(err.Error()) +"\"}"))
		}
		return
	}

	mapResponse := map[string]interface{}{"error": "", "response": res}

	data, _ := json.Marshal(mapResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
func (in *OtherCreateParams) ApiGenSet(r *http.Request) error {

	// Username


	UsernameRaw := r.URL.Query().Get("username")
	if r.Method == "POST" {
		UsernameRaw = r.FormValue("username")
	}

	if true && UsernameRaw == "" {
		return fmt.Errorf("username must me not empty")
	}

	if false && UsernameRaw == "" {
		UsernameRaw = ""
	}

	if false {
		isValid := false
		for _, enumValue := range strings.Split("", "|") {
			isValid = isValid || UsernameRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("username must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if true && len(UsernameRaw) < 3 {
		return fmt.Errorf("username len must be >= 3")
	}
	in.Username = UsernameRaw

	// Name


	NameRaw := r.URL.Query().Get("account_name")
	if r.Method == "POST" {
		NameRaw = r.FormValue("account_name")
	}

	if false && NameRaw == "" {
		return fmt.Errorf("account_name must me not empty")
	}

	if false && NameRaw == "" {
		NameRaw = ""
	}

	if false {
		isValid := false
		for _, enumValue := range strings.Split("", "|") {
			isValid = isValid || NameRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("account_name must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if false && len(NameRaw) < 0 {
		return fmt.Errorf("account_name len must be >= 0")
	}
	in.Name = NameRaw

	// Class


	ClassRaw := r.URL.Query().Get("class")
	if r.Method == "POST" {
		ClassRaw = r.FormValue("class")
	}

	if false && ClassRaw == "" {
		return fmt.Errorf("class must me not empty")
	}

	if true && ClassRaw == "" {
		ClassRaw = "warrior"
	}

	if true {
		isValid := false
		for _, enumValue := range strings.Split("warrior|sorcerer|rouge", "|") {
			isValid = isValid || ClassRaw == enumValue
		}
		if !isValid {
			return fmt.Errorf("class must be one of [" + strings.Join(strings.Split("warrior|sorcerer|rouge", "|"), ", ") + "]")
		}
	}

	if false && len(ClassRaw) < 0 {
		return fmt.Errorf("class len must be >= 0")
	}
	in.Class = ClassRaw

	// Level
	LevelRaw := r.URL.Query().Get("level")
	if r.Method == "POST" {
		LevelRaw = r.FormValue("level")
	}

	var LevelValue int

	if false && LevelRaw == "" {
		return fmt.Errorf("level must me not empty")
	}

	if LevelRaw != "" {
		LevelConvValue, LevelErr := strconv.Atoi(LevelRaw)
		if LevelErr !=nil {
			return fmt.Errorf("level must be int")
		}
		LevelValue = LevelConvValue
	}

	if false && LevelRaw == "" {
		if "" != "" {
			LevelValue, _ = strconv.Atoi("")
		}
		
	}

	if false {
		isValid := false

		for _, enumValue := range strings.Split("", "|") {
			enumIntValue, _ := strconv.Atoi(enumValue)
			isValid = isValid || LevelValue == enumIntValue
		}

		if !isValid {
			return fmt.Errorf("level must be one of [" + strings.Join(strings.Split("", "|"), ", ") + "]")
		}
	}

	if true && LevelValue < 1 {
		return fmt.Errorf("level must be >= 1")
	}

	if true && LevelValue > 50 {
		return fmt.Errorf("level must be <= 50")
	}

	in.Level = LevelValue
    return nil
}


func (api *OtherApi) wrapperApiGenCreate(w http.ResponseWriter, r *http.Request){
	if "POST" != "" && r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\" : \"bad method\"}"))
		return
	}

	if true && r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\" : \"unauthorized\"}"))
		return
	}

	var p OtherCreateParams
	setErr := p.ApiGenSet(r)
	if setErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + strings.ToLower(setErr.Error()) +"\"}"))
		return
	}
	
	res, err := api.Create(r.Context(), p)

	if err != nil {
		switch err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(apiErr.Err.Error()) +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(err.Error()) +"\"}"))
		}
		return
	}

	mapResponse := map[string]interface{}{"error": "", "response": res}

	data, _ := json.Marshal(mapResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/user/profile":
        api.wrapperApiGenProfile(w,r)
    case "/user/create":
        api.wrapperApiGenCreate(w,r)
    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte("{\"error\": \"unknown method\"}"))
    }
}
func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/user/create":
        api.wrapperApiGenCreate(w,r)
    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte("{\"error\": \"unknown method\"}"))
    }
}
