package main

import (
	"errors"
	"fmt"
	"net/http"
)

type (
	ZFSListRequest struct {
		Type   string `json:"type"`
		Prefix string `json:"prefix"`
	}

	ZFSGetRequest struct {
		Name string `json:"name"`
	}

	SetRequest struct {
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties,omitempty"`
	}
)

func (z *ZFS) List(r *http.Request, req *ZFSListRequest, resp *[]*Dataset) error {
	if req.Type == "" {
		req.Type = "all"
	}
	ds, err := listByType(req.Type, req.Prefix)
	if err != nil {
		return err
	}
	*resp = ds
	return nil
}

func (z *ZFS) Get(r *http.Request, req *ZFSGetRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	ds, err := getDataset(req.Name)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}

func (z *ZFS) Set(r *http.Request, req *SetRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}
	if req.Properties == nil || len(req.Properties) == 0 {
		return errors.New("must have properties")
	}
	ds, err := getDataset(req.Name)

	args := make([]string, 1, 2+(len(req.Properties)*2))
	args[0] = "set"

	for k, v := range req.Properties {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, req.Name)
	_, err = zfs(args...)
	if err != nil {
		return err
	}

	ds, err = getDataset(req.Name)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}
