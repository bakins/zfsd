package main

// based on https://github.com/mistifyio/go-zfs

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

type (
	command struct {
		Command string
		Stdin   io.Reader
		Stdout  io.Writer
	}

	// Dataset is a zfs dataset.  This could be a volume, filesystem, snapshot. Check the type field
	// The field definitions can be found in the zfs manual: http://www.freebsd.org/cgi/man.cgi?zfs(8)
	Dataset struct {
		Name        string `json:"name"`
		Used        uint64 `json:"used,omitempty"`
		Available   uint64 `json:"available,omitempty"`
		Mountpoint  string `json:"mountpoint,omitempty"`
		Compression string `json:"compression,omitempty"`
		Type        string `json:"type"`
		Written     uint64 `json:"written,omitempty"`
		Volsize     uint64 `json:"volsize,omitempty"`
		Quota       uint64 `json:"quota,omitempty"`
		Origin      string `json:"origin,omitempty"`
	}
)

// helper function to wrap typical calls to zfs
func zfs(arg ...string) ([][]string, error) {
	c := command{Command: "zfs"}
	return c.Run(arg...)
}

func (c *command) Run(arg ...string) ([][]string, error) {

	cmd := exec.Command(c.Command, arg...)

	var stdout, stderr bytes.Buffer

	if c.Stdout == nil {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = c.Stdout
	}

	if c.Stdin != nil {
		cmd.Stdin = c.Stdin

	}
	cmd.Stderr = &stderr

	debug := strings.Join([]string{cmd.Path, strings.Join(cmd.Args, " ")}, " ")
	fmt.Println(debug)
	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("%s: '%s' => %s", err, debug, stderr.String())
	}

	// assume if you passed in something for stdout, that you know what to do with it
	if c.Stdout != nil {
		return nil, nil
	}

	lines := strings.Split(stdout.String(), "\n")

	//last line is always blank
	lines = lines[0 : len(lines)-1]
	output := make([][]string, len(lines))

	for i, l := range lines {
		output[i] = strings.Split(l, "\t")
	}

	return output, nil
}

var propertyFields = make([]string, 0, 66)
var propertyMap = map[string]string{}

func init() {
	st := reflect.TypeOf(Dataset{})
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		// only look at exported values
		if f.PkgPath == "" {
			key := strings.ToLower(f.Name)
			propertyMap[key] = f.Name
			propertyFields = append(propertyFields, key)
		}
	}

}

func parseDatasetLine(line []string) (*Dataset, error) {
	dataset := Dataset{}

	st := reflect.ValueOf(&dataset).Elem()

	for j, field := range propertyFields {

		fieldName := propertyMap[field]

		if fieldName == "" {
			continue
		}

		f := st.FieldByName(fieldName)
		value := line[j]

		switch f.Kind() {

		case reflect.Uint64:
			var v uint64
			if value != "-" {
				v, _ = strconv.ParseUint(value, 10, 64)
			}
			f.SetUint(v)

		case reflect.String:
			v := ""
			if value != "-" {
				v = value
			}
			f.SetString(v)

		}
	}
	return &dataset, nil
}

func parseDatasetLines(lines [][]string) ([]*Dataset, error) {
	datasets := make([]*Dataset, len(lines))

	for i, line := range lines {
		d, _ := parseDatasetLine(line)
		datasets[i] = d
	}

	return datasets, nil
}

func listByType(t, filter string) ([]*Dataset, error) {
	args := []string{"list", "-t", t, "-rHpo", strings.Join(propertyFields, ",")}[:]
	if filter != "" {
		args = append(args, filter)
	}
	out, err := zfs(args...)
	if err != nil {
		return nil, err
	}
	return parseDatasetLines(out)
}

func propsSlice(properties map[string]string) []string {
	args := make([]string, 0, len(properties)*3)
	for k, v := range properties {
		args = append(args, "-o")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

// GetDataset retrieves a single dataset
func getDataset(name string) (*Dataset, error) {
	out, err := zfs("list", "-Hpo", strings.Join(propertyFields, ","), name)
	if err != nil {
		return nil, err
	}
	return parseDatasetLine(out[0])
}
