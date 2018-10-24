package main

import (
	"errors"
	"strconv"
	"strings"
	"reflect"

	"os/user"
	"io/ioutil"

	"github.com/fatih/structtag"
	"gopkg.in/src-d/go-git.v4/config"
	c "gopkg.in/src-d/go-git.v4/plumbing/format/config"
)

type GitConfig struct {

}

func (g *GitConfig) ParseFile(s interface{}) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	raw, err := ioutil.ReadFile(usr.HomeDir + "/.gitconfig")
	if err != nil {
		return err
	}
	return g.Parse(raw, s)
}

func (g *GitConfig) parseOptions(fld *reflect.Value, opts c.Options) error {
	for _, opt := range opts {
		optField := fld.FieldByName(strings.Title(opt.Key))
		switch optType :=  optField.Type().String(); optType {
		case "bool": {
			if optField.CanAddr() && optField.CanSet() {
				if opt.Value == "true" || opt.Value == "1" {
					optField.SetBool(true)
				}
			}
		}
		case "string": {
			if optField.CanAddr() && optField.CanSet() {
				optField.SetString(opt.Value)
			}
		}
		case "int64": {
			if optField.CanAddr() && optField.CanSet() {
				v, _ := strconv.ParseInt(opt.Value, 10, 64)
				optField.SetInt(v)
			}
		}
		case "float64": {
			if optField.CanAddr() && optField.CanSet() {
				v, _ := strconv.ParseFloat(opt.Value, 64)
				optField.SetFloat(v)
			}
		}
		default:
			return errors.New("Unsupported type " + optType)
		}
	}
	return nil
}

func (g *GitConfig) Parse(raw []byte, s interface{}) error {
	elem := reflect.ValueOf(s).Elem()
	flds := elem.Type()
	c := config.Config{}
	for i := 0; i < flds.NumField(); i++ {
		f := flds.Field(i)
		fld := elem.FieldByName(f.Name)
		tag := f.Tag
		tags, err := structtag.Parse(string(tag))
		if err != nil {
			panic(err)
		}
		t, err := tags.Get("git")
		var section string
		if err == nil {
			section = t.Name
		} else {
			section = strings.ToLower(f.Name)
		}

		
		if err := c.Unmarshal(raw); err != nil {
			return err
		}
		for _, sub := range c.Raw.Section(section).Subsections {
			subField := fld.FieldByName(strings.Title(sub.Name))
			g.parseOptions(&subField, sub.Options)
		}
		g.parseOptions(&fld, c.Raw.Section(section).Options)
	}
	return nil
}