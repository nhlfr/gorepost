// Copyright 2015 Robert S. Gerus. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bot

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xpath"

	cfg "github.com/arachnist/gorepost/config"
	"github.com/arachnist/gorepost/irc"
)

var elementNotFound = errors.New("Element not found in document")

func httpGet(l string) ([]byte, error) {
	var buf []byte
	tr := &http.Transport{
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(l)
	if err != nil {
		return []byte{}, err
	}

	// 5MiB
	if resp.ContentLength > 5*1024*1024 || resp.ContentLength < 0 {
		buf = make([]byte, 5*1024*1024)
	} else if resp.ContentLength == 0 {
		return []byte{}, nil
	} else {
		buf = make([]byte, resp.ContentLength)
	}

	i, err := io.ReadFull(resp.Body, buf)
	if err == io.ErrUnexpectedEOF {
		buf = buf[:i]
	} else if err != nil {
		return []byte{}, err
	}

	return buf, nil
}

func httpGetXpath(l, x string) (string, error) {
	buf, err := httpGet(l)
	if err != nil {
		return "", err
	}

	doc, err := gokogiri.ParseHtml(buf)
	defer doc.Free()
	if err != nil {
		return "", err
	}

	xpath := xpath.Compile(x)
	sr, err := doc.Root().Search(xpath)
	if err != nil {
		return "", err
	}

	if len(sr) > 0 {
		return sr[0].InnerHtml(), nil
	} else {
		return "", elementNotFound
	}
}

func reply(msg irc.Message, text string) irc.Message {
	if msg.Params[0] == cfg.LookupString(msg.Context, "Nick") {
		return irc.Message{
			Command:  "PRIVMSG",
			Params:   []string{msg.Prefix.Name},
			Trailing: text,
		}
	} else {
		return irc.Message{
			Command:  "PRIVMSG",
			Params:   msg.Params,
			Trailing: text,
		}
	}
}