// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
)

func (db *Lowlevel) AddOrUpdatePendingDevice(device protocol.DeviceID, name, address string) error {
	key := db.keyer.GeneratePendingDeviceKey(nil, device[:])
	od := ObservedDevice{
		Time:    time.Now().Round(time.Second),
		Name:    name,
		Address: address,
	}
	bs, err := od.Marshal()
	if err == nil {
		err = db.Put(key, bs)
	}
	return err
}

func (db *Lowlevel) RemovePendingDevice(device protocol.DeviceID) {
	key := db.keyer.GeneratePendingDeviceKey(nil, device[:])
	if err := db.Delete(key); err != nil {
		l.Warnf("Failed to remove pending device entry: %v", err)
	}
}

// PendingDevice combines the DB key and value contents for passing around
type PendingDevice struct {
	DeviceID protocol.DeviceID
	ObservedDevice
}

// PendingDevices drops any invalid entries from the database after a
// warning log message, as a side-effect.  That's the only possible
// "repair" measure and appropriate for the importance of pending
// entries.  They will come back soon if still relevant.
func (db *Lowlevel) PendingDevices() ([]PendingDevice, error) {
	iter, err := db.NewPrefixIterator([]byte{KeyTypePendingDevice})
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	var res []PendingDevice
	for iter.Next() {
		keyDev := db.keyer.DeviceFromPendingDeviceKey(iter.Key())
		deviceID, err := protocol.DeviceIDFromBytes(keyDev)
		var bs []byte
		var od ObservedDevice
		if err != nil {
			goto deleteKey
		}
		if bs, err = db.Get(iter.Key()); err != nil {
			goto deleteKey
		}
		if err = od.Unmarshal(bs); err != nil {
			goto deleteKey
		}
		res = append(res, PendingDevice{deviceID, od})
		continue
	deleteKey:
		l.Infof("Invalid pending device entry, deleting from database: %x", iter.Key())
		if err := db.Delete(iter.Key()); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (db *Lowlevel) AddOrUpdatePendingFolder(id, label string, device protocol.DeviceID) error {
	key, err := db.keyer.GeneratePendingFolderKey(nil, device[:], []byte(id))
	if err != nil {
		return err
	}
	of := ObservedFolder{
		Time:  time.Now().Round(time.Second),
		Label: label,
	}
	bs, err := of.Marshal()
	if err == nil {
		err = db.Put(key, bs)
	}
	return err
}

// RemovePendingFolder removes entries for specific folder / device combinations
func (db *Lowlevel) RemovePendingFolder(pf PendingFolder) {
	key, err := db.keyer.GeneratePendingFolderKey(nil, []byte(pf.FolderID), pf.DeviceID[:])
	if err == nil {
		if err = db.Delete(key); err == nil {
			return
		}
	}
	l.Warnf("Failed to remove pending folder entry: %v", err)
}

// PendingFolder combines the DB key and value contents for passing around
type PendingFolder struct {
	FolderID string
	DeviceID protocol.DeviceID
	ObservedFolder
}

// PendingFolders drops any invalid entries from the database as a side-effect.
func (db *Lowlevel) PendingFolders() ([]PendingFolder, error) {
	iter, err := db.NewPrefixIterator([]byte{KeyTypePendingFolder})
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	var res []PendingFolder
	for iter.Next() {
		keyDev, ok := db.keyer.DeviceFromPendingFolderKey(iter.Key())
		deviceID, err := protocol.DeviceIDFromBytes(keyDev)
		var of ObservedFolder
		var folderID string
		var bs []byte
		if !ok || err != nil {
			goto deleteKey
		}
		if folderID = string(db.keyer.FolderFromPendingFolderKey(iter.Key())); len(folderID) < 1 {
			goto deleteKey
		}
		if bs, err = db.Get(iter.Key()); err != nil {
			goto deleteKey
		}
		if err = of.Unmarshal(bs); err != nil {
			goto deleteKey
		}
		res = append(res, PendingFolder{folderID, deviceID, of})
		continue
	deleteKey:
		l.Infof("Invalid pending folder entry, deleting from database: %x", iter.Key())
		if err := db.Delete(iter.Key()); err != nil {
			return nil, err
		}
	}
	return res, nil
}
