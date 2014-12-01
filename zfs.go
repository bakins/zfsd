// Package main implements a simple HTTP interface for zfs management
package zfsd

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/mistifyio/go-zfs.v1"
)

type (

	// ZFS is used for RPC services
	ZFS struct {
	}

	ListRequest struct {
		Type   string `json:"type"`
		Prefix string `json:"prefix"`
	}

	GetRequest struct {
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

// List retrieves a list of all ZFS datasets, optionally only of a certain type or prefix
func (z *ZFS) List(r *http.Request, req *ListRequest, resp *[]*zfs.Dataset) error {

	var ds []*zfs.Dataset
	var err error
	switch req.Type {
	case "snapshot":
		ds, err = zfs.Snapshots(req.Prefix)
	case "filesystem":
		ds, err = zfs.Filesystems(req.Prefix)
	case "volume":
		ds, err = zfs.Volumes(req.Prefix)
	case "", "all":
		ds, err = zfs.Datasets(req.Prefix)
	default:
		fmt.Errorf("unknown type: %s", req.Type)
	}
	if err != nil {
		return err
	}
	*resp = ds
	return nil
}

// Get retrieves a single ZFS dataset.
func (z *ZFS) Get(r *http.Request, req *GetRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}

// Set sets properties on a ZFS dataset.
func (z *ZFS) Set(r *http.Request, req *SetRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}
	if req.Properties == nil || len(req.Properties) == 0 {
		return errors.New("must have properties")
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	// zfs should have a setproperties that takes a map
	for k, v := range req.Properties {
		err := ds.SetProperty(k, v)
		if err != nil {
			return err
		}
	}
	ds, err = zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}

// Snapshot creates a snapshot of a ZFS dataset.
func (z *ZFS) Snapshot(r *http.Request, req *SnapshotRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	snap, err := ds.Snapshot(req.Snapshot, false)
	if err != nil {
		return err
	}

	*resp = *snap
	return nil
}

// Clone clones a ZFS snapshot
func (z *ZFS) Clone(r *http.Request, req *CloneRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	if req.Target == "" {
		return errors.New("must have target")
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	clone, err := ds.Clone(req.Target, req.Properties)
	if err != nil {
		return err
	}

	*resp = *clone
	return nil
}

// Destroy removes a ZFS dataset
func (z *ZFS) Destroy(r *http.Request, req *DestroyRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	err = ds.Destroy(req.Recursive)
	if err != nil {
		return err
	}
	*resp = *ds
	return nil
}

// Rollback rolls the given dataset to back a previous snapshot.
func (z *ZFS) Rollback(r *http.Request, req *RollbackRequest, resp *zfs.Dataset) error {
	if req.Name == "" {
		return errors.New("must have name")
	}

	if req.Snapshot == "" {
		return errors.New("must have snapshot")
	}

	snapName := fmt.Sprintf("%s@%s", req.Name, req.Snapshot)
	snap, err := zfs.GetDataset(snapName)
	if err != nil {
		return err
	}

	err = snap.Rollback(false)
	if err != nil {
		return err
	}

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	*resp = *ds
	return nil
}
