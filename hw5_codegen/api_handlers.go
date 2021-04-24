package main


func (api *MyApi) wrapperApiGenProfile(w http.ResponseWriter, r *http.Request){
	if r.Method != "" {
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
	p.fillApiGen(r)
	validateErr := p.validateApiGen()
	if validateErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + validateErr.Error() +"\"}"))
		return
	}
	
	res, err := h.Profile(r.Context(), p)

	if err != nil {
		switch v := err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + apiErr.Err.Error() +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + err.Error() +"\"}"))
		}
		return
	}

	data, _ := json.Marshal(res)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (api *MyApi) wrapperApiGenCreate(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
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
	p.fillApiGen(r)
	validateErr := p.validateApiGen()
	if validateErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + validateErr.Error() +"\"}"))
		return
	}
	
	res, err := h.Create(r.Context(), p)

	if err != nil {
		switch v := err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + apiErr.Err.Error() +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + err.Error() +"\"}"))
		}
		return
	}

	data, _ := json.Marshal(res)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (api *OtherApi) wrapperApiGenCreate(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
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
	p.fillApiGen(r)
	validateErr := p.validateApiGen()
	if validateErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + validateErr.Error() +"\"}"))
		return
	}
	
	res, err := h.Create(r.Context(), p)

	if err != nil {
		switch v := err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + apiErr.Err.Error() +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + err.Error() +"\"}"))
		}
		return
	}

	data, _ := json.Marshal(res)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
