// Copyright 2024 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tsuru/go-tsuruclient/pkg/config"
	tsuruHTTP "github.com/tsuru/tsuru-client/tsuru/http"
	"github.com/tsuru/tsuru/cmd"
)

func nativeLogin(ctx *cmd.Context) error {
	var email string

	// Use raw output to avoid missing the input prompt messages
	ctx.RawOutput()

	if len(ctx.Args) > 0 {
		email = ctx.Args[0]
	} else {
		fmt.Fprint(ctx.Stdout, "Email: ")
		fmt.Fscanf(ctx.Stdin, "%s\n", &email)
	}
	fmt.Fprint(ctx.Stdout, "Password: ")
	password, err := cmd.PasswordFromReader(ctx.Stdin)
	if err != nil {
		return err
	}
	fmt.Fprintln(ctx.Stdout)
	u, err := config.GetURL("/users/" + email + "/tokens")
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Set("password", password)
	b := strings.NewReader(v.Encode())
	request, err := http.NewRequest("POST", u, b)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := tsuruHTTP.UnauthenticatedClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	out := make(map[string]interface{})
	err = json.Unmarshal(result, &out)
	if err != nil {
		return err
	}
	fmt.Fprintln(ctx.Stdout, "Successfully logged in!")
	err = config.RemoveTokenV2()
	if err != nil {
		return err
	}
	return config.WriteTokenV1(out["token"].(string))
}
