package main

import (
	"errors"
	"fmt"
	"net/http"
)

type (
	SnapshotRequest struct {
		Name     string `json:"name"`
		Snapshot string `json:"snapshot"`
	}

	CloneRequest struct {
		Name       string            `json:"name"`
		Snapshot   string            `json:"snapshot"`
		Target     string            `json:"target"`
		Properties map[string]string `json:"properties,omitempty"`
	}
)

func (z *ZFS) Snapshot(r *http.Request, req *SnapshotRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	ds, err := getDataset(req.Name)
	if err != nil {
		return err
	}

	args := make([]string, 1, 4)
	args[0] = "snapshot"

	snapName := fmt.Sprintf("%s@%s", ds.Name, req.Snapshot)
	args = append(args, snapName)
	_, err = zfs(args...)
	if err != nil {
		return err
	}
	snap, err := getDataset(snapName)
	if err != nil {
		return err
	}

	*resp = *snap
	return nil
}

func (z *ZFS) Clone(r *http.Request, req *CloneRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	if req.Target == "" {
		return errors.New("must have target")
	}

	snapName := fmt.Sprintf("%s@%s", req.Name, req.Snapshot)
	snap, err := getDataset(snapName)
	if err != nil {
		return err
	}

	if snap.Type != "snapshot" {
		return errors.New("not a snapshot")
	}

	args := make([]string, 1, 4)
	args[0] = "clone"
	if req.Properties != nil {
		args = append(args, propsSlice(req.Properties)...)
	}
	args = append(args, []string{snap.Name, req.Target}...)
	_, err = zfs(args...)
	if err != nil {
		return err
	}

	ds, err := getDataset(req.Target)
	if err != nil {
		return err
	}

	*resp = *ds
	return nil
}
