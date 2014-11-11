package main

import (
	"errors"
	"fmt"
	"net/http"
)

type (

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

	SnapshotRequest struct {
		Name     string `json:"name"`
		Snapshot string `json:"snapshot"`
	}

	RollbackRequest struct {
		Name      string `json:"name"`
		Snapshot  string `json:"snapshot"`
		Recursive bool   `json:"recursive"`
	}

	CloneRequest struct {
		Name       string            `json:"name"`
		Snapshot   string            `json:"snapshot"`
		Target     string            `json:"target"`
		Properties map[string]string `json:"properties,omitempty"`
	}

	DestroyRequest struct {
		Name      string `json:"name"`
		Recursive bool   `json:"recursive"`
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

	args := make([]string, 1, 4)
	args[0] = "clone"
	if req.Properties != nil {
		args = append(args, propsSlice(req.Properties)...)
	}
	args = append(args, []string{req.Name, req.Target}...)
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

func (z *ZFS) Destroy(r *http.Request, req *DestroyRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
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

func (z *ZFS) Rollback(r *http.Request, req *RollbackRequest, resp *Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	snapName := fmt.Sprintf("%s@%s", req.Name, req.Snapshot)
	snap, err := getDataset(snapName)
	if err != nil {
		return err
	}

	if snap.Type != "snapshot" {
		return errors.New("not a snapshot")
	}

	args := make([]string, 1, 3)
	args[0] = "rollback"
	if req.Recursive {
		args = append(args, "-r")
	}
	args = append(args, snapName)

	_, err = zfs(args...)
	if err != nil {
		return err
	}

	ds, err := getDataset(req.Name)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}
