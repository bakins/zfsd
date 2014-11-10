package main

import (
	"errors"
	"net/http"
)

type (
	DestroyRequest struct {
		Name      string `json:"name"`
		Recursive bool   `json:"recursive"`
	}
)

func (z *ZFS) Destroy(r *http.Request, req *DestroyRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	ds, err := getDataset(req.Name)
	if err != nil {
		return err
	}

	args := make([]string, 1, 3)
	args[0] = "destroy"
	if req.Recursive {
		args = append(args, "-r")
	}
	args = append(args, ds.Name)

	_, err = zfs(args...)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}
