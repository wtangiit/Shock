// Package globus implements MG-RAST OAuth authentication
package mgrast

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/user"
	client "github.com/jaredwilkening/httpclient"
	"io/ioutil"
	"strconv"
	"strings"
)

type resErr struct {
	error string `json:"error"`
}

type credentials struct {
	Uname  string   `json:"user"`
	Fname  string   `json:"firstname"`
	Lname  string   `json:"lastname"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

func authHeaderType(header string) string {
	tmp := strings.Split(header, " ")
	if len(tmp) > 1 {
		return strings.ToLower(tmp[0])
	}
	return ""
}

// Auth takes the request authorization header and returns
// user
func Auth(header string) (*user.User, error) {
	switch authHeaderType(header) {
	case "mgrast", "oauth":
		return authToken(strings.Split(header, " ")[1])
	case "basic":
		return nil, errors.New("This authentication method does not support username/password authentication. Please use MG-RAST your token.")
	default:
		return nil, errors.New("Invalid authentication header.")
	}
}

// authToken validiates token by fetching user information.
func authToken(t string) (*user.User, error) {
	url := conf.Conf["mgrast_oauth_url"]
	if url == "" {
		return nil, errors.New("mgrast_oauth_url not set in configuration")
	}

	form := client.NewForm()
	form.AddParam("token", t)
	form.AddParam("action", "credentials")
	form.AddParam("groups", "true")
	err := form.Create()
	if err != nil {
		return nil, err
	}

	headers := client.Header{
		"Content-Type":   form.ContentType,
		"Content-Length": strconv.FormatInt(form.Length, 10),
	}

	if res, err := client.Do("POST", url, headers, form.Reader); err == nil {
		if res.StatusCode == 200 {
			r := credentials{}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err != nil {
				return nil, err
			}
			return &user.User{Username: r.Uname, Fullname: r.Fname + " " + r.Lname, Email: r.Email, CustomFields: map[string][]string{"groups": r.Groups}}, nil
		} else {
			r := resErr{}
			body, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("%s\n", body)
			if err = json.Unmarshal(body, &r); err == nil {
				return nil, errors.New("request error: " + res.Status)
			} else {
				return nil, errors.New(res.Status + ": " + r.error)
			}
		}
	}
	return nil, nil
}
