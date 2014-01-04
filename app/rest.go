package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"ext/go-json-rest"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// supporting functions for a restfull app server

// interface for all inbound json payloads
type GeneralRequest interface {
	Token() string
}

// a minimal request body - perjhaps for authenticated get requests
type MinRequest struct {
	Key string `json:"token"`
}

func (a MinRequest) Token() string {
	return a.Key
}

// standard response wrapper
type RestResponse struct {
	Token   string      `json:"token"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Obj     interface{} `json:"obj"`
}

var emailRegex *regexp.Regexp
var noneAlphaNum *regexp.Regexp

var validID = regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)

func PrepRest() {
	emailRegex = regexp.MustCompile(`.+@.+\..+`)
	noneAlphaNum = regexp.MustCompile(`\W`)
}

func ValidateAndTidyEmail(email string) (string, bool) {
	tidy := TidyEmail(email)
	return tidy, emailRegex.MatchString(tidy)
}

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func TidyEmail(fld string) string {
	return strings.ToLower(strings.Replace(fld, " ", "", -1))
}

func RemoveUtf8AndMultipleSpaces(fld string) string {
	bits := []string{}
	for _, s := range strings.Split(fld, " ") {
		s = noneAlphaNum.ReplaceAllString(s, "")
		if s != "" {
			bits = append(bits, s)
		}
	}

	return strings.Join(bits, " ")
}

// remove multiple spaces, lower case it, strip leading and trailing, and replace single spaces with '-'
func TidyKey(fld string) string {
	fld = strings.Replace(fld, ".", "_", -1)
	fld = strings.Replace(fld, "-", "_", -1)

	return strings.TrimSpace(strings.ToLower(strings.Replace(RemoveUtf8AndMultipleSpaces(fld), " ", "_", -1)))
}

func TidyInput(fld string) string {
	return strings.TrimSpace(fld)
}

// return the key from the url
func getKey(req *rest.Request) (string, bool) {
	kl, in := req.URL.Query()["key"]
	if !in {
		return "", false
	}
	return kl[0], true
}

// given a request and general payload and if check authorisation
// find a session token(key) in the url or payload return a session, and http code and error
func PrepRequest(req *rest.Request, obj GeneralRequest, needAuth bool) (*Sesh, int, error) {

	plToken := ""

	err := req.DecodeJsonPayload(obj)
	if err == nil {
		plToken = obj.Token()
	}

	key, in := getKey(req)
	if in {
		if plToken != "" {
			if plToken != key {
				return nil, http.StatusBadRequest, errors.New("key in the url conflicts with the key in the body")
			}
		}
	} else {
		key = plToken
	}

	if key == "" {
		return nil, http.StatusBadRequest, errors.New("No key found in the session")
	}

	sesh, found, err := GetSeshFromCache(key)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if needAuth {
		authed := true
		if !found {
			authed = false
		} else if !sesh.Auth {
			authed = false
		}
		if !authed {
			return nil, http.StatusUnauthorized, errors.New("unauthorised")
		}
	}

	// got no valid session so make one
	if sesh == nil {
		sesh = NewSesh()
	}

	return sesh, 200, nil
}

func RestPost(url string, obj interface{}, repl interface{}) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	} else {

		body := bytes.NewBuffer(b)

		resp, err := http.Post(url, "application/json", body)

		if err != nil {
			return err
		} else {
			defer resp.Body.Close()
			if repl != nil {

				content, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					return err
				} else {
					err := json.Unmarshal(content, repl)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
