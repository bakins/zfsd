package main

import (
	"errors"
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
