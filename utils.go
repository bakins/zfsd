package zfsd

// based on https://github.com/mistifyio/go-zfs

// TODO: investigate getting rid of all the reflection

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type (
	command struct {
		Command string
		Stdin   io.Reader
		Stdout  io.Writer
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
		return nil, fmt.Errorf(stderr.String())
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

func listByType(t, filter string) ([]*Dataset, error) {
	args := []string{"get", "all", "-t", t, "-rHp"}
	if filter != "" {
		args = append(args, filter)
	}
	out, err := zfs(args...)
	if err != nil {
		return nil, err
	}

	datasets := make([]*Dataset, 0)
	name := ""
	var ds *Dataset
	for _, line := range out {
		if name != line[0] {
			name = line[0]
			ds = &Dataset{Name: name}
			datasets = append(datasets, ds)
		}
		ds.parseLine(line)
	}

	return datasets, nil
}

func propsSlice(properties map[string]string) []string {
	args := make([]string, 0, len(properties)*3)
	for k, v := range properties {
		args = append(args, "-o")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func setString(field *string, value string) {
	v := ""
	if value != "-" {
		v = value
	}
	*field = v
}

func setUint(field *uint64, value string) {
	var v uint64
	if value != "-" {
		v, _ = strconv.ParseUint(value, 10, 64)
	}
	*field = v
}

func (ds *Dataset) parseLine(line []string) {
	prop := line[1]
	val := line[2]

	switch prop {
	case "available":
		setUint(&ds.Available, val)
	case "compression":
		setString(&ds.Compression, val)
	case "mountpoint":
		setString(&ds.Mountpoint, val)
	case "quota":
		setUint(&ds.Quota, val)
	case "type":
		setString(&ds.Type, val)
	case "used":
		setUint(&ds.Used, val)
	case "volsize":
		setUint(&ds.Volsize, val)
	case "written":
		setUint(&ds.Written, val)
	}
}

// GetDataset retrieves a single dataset
func getDataset(name string) (*Dataset, error) {
	out, err := zfs("get", "all", "-Hp", name)
	if err != nil {
		return nil, err
	}

	ds := &Dataset{Name: name}
	for _, line := range out {
		ds.parseLine(line)
	}

	return ds, nil
}
