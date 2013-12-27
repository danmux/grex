package app

import (
	"errors"
	"ext/go-json-rest"
	"net/http"
	"regexp"
	"strings"
)

// supporting functions for a restfull app server

// interface for all inbound json payloads
type GeneralRequest interface {
	Token() string
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

func RemoveUtf8AndultipleSpaces(fld string) string {
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

	return strings.TrimSpace(strings.ToLower(strings.Replace(RemoveUtf8AndultipleSpaces(fld), " ", "-", -1)))
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
	key, in := getKey(req)

	err := req.DecodeJsonPayload(obj)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	plToken := obj.Token()
	if in {
		if plToken != "" {
			if plToken != key {
				return nil, http.StatusBadRequest, errors.New("key in the url conflicts with the key in the body")
			}
		}
	} else {
		key = plToken
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
