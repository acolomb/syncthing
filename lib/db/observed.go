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
	if err != nil {
		return err
	}
	return db.Put(key, bs)
}

func (db *Lowlevel) RemovePendingDevice(device protocol.DeviceID) {
	key := db.keyer.GeneratePendingDeviceKey(nil, device[:])
	if err := db.Delete(key); err != nil {
		l.Warnf("Failed to remove pending device entry: %v", err)
	}
}

// PendingDevices enumerates all entries.  Invalid ones are dropped from the database
// after a warning log message, as a side-effect.
func (db *Lowlevel) PendingDevices() (map[protocol.DeviceID]ObservedDevice, error) {
	iter, err := db.NewPrefixIterator([]byte{KeyTypePendingDevice})
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	res := make(map[protocol.DeviceID]ObservedDevice)
	for iter.Next() {
		keyDev := db.keyer.DeviceFromPendingDeviceKey(iter.Key())
		deviceID, err := protocol.DeviceIDFromBytes(keyDev)
		var od ObservedDevice
		if err != nil {
			goto deleteKey
		}
		if err = od.Unmarshal(iter.Value()); err != nil {
			goto deleteKey
		}
		res[deviceID] = od
		continue
	deleteKey:
		// Deleting invalid entries is the only possible "repair" measure and
		// appropriate for the importance of pending entries.  They will come back
		// soon if still relevant.
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
	if err != nil {
		return err
	}
	return db.Put(key, bs)
}

// RemovePendingFolderForDevice removes entries for specific folder / device combinations.
func (db *Lowlevel) RemovePendingFolderForDevice(id string, device protocol.DeviceID) {
	key, err := db.keyer.GeneratePendingFolderKey(nil, device[:], []byte(id))
	if err != nil {
		return
	}
	if err := db.Delete(key); err != nil {
		l.Warnf("Failed to remove pending folder entry: %v", err)
	}
}

// RemovePendingFolder removes all entries matching a specific folder ID.
func (db *Lowlevel) RemovePendingFolder(id string) {
	iter, err := db.NewPrefixIterator([]byte{KeyTypePendingFolder})
	if err != nil {
		l.Infof("Could not iterate through pending folder entries: %v", err)
		return
	}
	defer iter.Release()
	for iter.Next() {
		if id != string(db.keyer.FolderFromPendingFolderKey(iter.Key())) {
			continue
		}
		if err := db.Delete(iter.Key()); err != nil {
			l.Warnf("Failed to remove pending folder entry: %v", err)
		}
	}
}

// Consolidated information about a pending folder
type PendingFolder struct {
	OfferedBy map[protocol.DeviceID]ObservedFolder `json:"offeredBy"`
}

func (db *Lowlevel) PendingFolders() (map[string]PendingFolder, error) {
	return db.PendingFoldersForDevice(protocol.EmptyDeviceID)
}

// PendingFoldersForDevice enumerates only entries matching the given device ID, unless it
// is EmptyDeviceID.  Invalid ones are dropped from the database after a warning log
// message, as a side-effect.
func (db *Lowlevel) PendingFoldersForDevice(device protocol.DeviceID) (map[string]PendingFolder, error) {
	var err error
	prefixKey := []byte{KeyTypePendingFolder}
	if device != protocol.EmptyDeviceID {
		prefixKey, err = db.keyer.GeneratePendingFolderKey(nil, device[:], nil)
		if err != nil {
			return nil, err
		}
	}
	iter, err := db.NewPrefixIterator(prefixKey)
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	res := make(map[string]PendingFolder)
	for iter.Next() {
		keyDev, ok := db.keyer.DeviceFromPendingFolderKey(iter.Key())
		deviceID, err := protocol.DeviceIDFromBytes(keyDev)
		var of ObservedFolder
		var folderID string
		if !ok || err != nil {
			goto deleteKey
		}
		if folderID = string(db.keyer.FolderFromPendingFolderKey(iter.Key())); len(folderID) < 1 {
			goto deleteKey
		}
		if err = of.Unmarshal(iter.Value()); err != nil {
			goto deleteKey
		}
		if _, ok := res[folderID]; !ok {
			res[folderID] = PendingFolder{
				OfferedBy: map[protocol.DeviceID]ObservedFolder{},
			}
		}
		res[folderID].OfferedBy[deviceID] = of
		continue
	deleteKey:
		// Deleting invalid entries is the only possible "repair" measure and
		// appropriate for the importance of pending entries.  They will come back
		// soon if still relevant.
		l.Infof("Invalid pending folder entry, deleting from database: %x", iter.Key())
		if err := db.Delete(iter.Key()); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (db *Lowlevel) AddOrUpdateCandidateLink(folder, label string, device, introducer protocol.DeviceID, meta *IntroducedDeviceDetails) error {
	key, err := db.keyer.GenerateCandidateLinkKey(nil, introducer[:], []byte(folder), device[:])
	if err != nil {
		return err
	}
	link := ObservedCandidateLink{
		//CommonFolder: ObservedFolder{
		Time:            time.Now().Round(time.Second),
		IntroducerLabel: label,
		//},
		CandidateMeta: meta,
	}
	bs, err := link.Marshal()
	if err == nil {
		err = db.Put(key, bs)
	}
	return err
}

// Details of a candidate device introduced through a specific folder:
// "Introducer says FolderID exists on device CandidateID"
type CandidateLink struct {
	Introducer  protocol.DeviceID
	FolderID    string
	CandidateID protocol.DeviceID

	ObservedCandidateLink //FIXME: Not needed if this granular info will only be used in cleanup!
}

func (db *Lowlevel) CandidateLinks() ([]CandidateLink, error) {
	res := []CandidateLink{
		{
			Introducer:  protocol.TestDeviceID2,
			FolderID:    "boggl-goggl",
			CandidateID: protocol.TestDeviceID1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "frob"}},
		{
			Introducer:  protocol.TestDeviceID2,
			FolderID:    "sleep-wells",
			CandidateID: protocol.TestDeviceID1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "nic"}},
		{
			Introducer:  protocol.TestDeviceID2,
			FolderID:    "damtn-omola",
			CandidateID: protocol.TestDeviceID1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "ate",
				CandidateMeta: &IntroducedDeviceDetails{
					CertName:      "foo",
					Addresses:     []string{"bar", "baz"},
					SuggestedName: "bazoo"},
			},
		},
	}
	return res, nil
}

// Consolidated information about a candidate device, enough to add a connection to it
type CandidateDevice struct {
	CertName     string                                     `json:"certName,omitempty"`
	Addresses    []string                                   `json:"addresses,omitempty"`
	IntroducedBy map[protocol.DeviceID]candidateAttribution `json:"introducedBy"`
}

// Details which an introducer told us about a candidate device
type candidateAttribution struct {
	Time          time.Time         `json:"time"`
	CommonFolders map[string]string `json:"commonFolders"`
	SuggestedName string            `json:"suggestedName,omitempty"`
}

func (db *Lowlevel) CandidateDevices(folder string) (map[protocol.DeviceID]CandidateDevice, error) {
	res := make(map[protocol.DeviceID]CandidateDevice)
	res[protocol.TestDeviceID1] = CandidateDevice{
		IntroducedBy: map[protocol.DeviceID]candidateAttribution{
			protocol.TestDeviceID2: candidateAttribution{
				// Should be the same for all folders, as they were all
				// mentioned in the most recent ClusterConfig
				Time: time.Now(),
				CommonFolders: map[string]string{
					"frob": "FROBBY",
					"nic":  "NICKY",
					"ate":  "ATEY",
				},
				// Only if the device ID is not known locally:
				SuggestedName: "bazoo",
			},
		},
		// Only if the device ID is not known locally:
		CertName:  "syncthing",
		Addresses: []string{"bar", "baz"},
	}
	res[protocol.TestDeviceID2] = CandidateDevice{
		IntroducedBy: map[protocol.DeviceID]candidateAttribution{
			protocol.TestDeviceID1: candidateAttribution{
				Time: time.Now(),
				CommonFolders: map[string]string{
					"dodo": "DODODODO",
				},
				SuggestedName: "coolnhip",
			},
		},
		CertName:  "syncthing",
		Addresses: []string{"bar", "baz"},
	}
	return res, nil
}

// Collect addresses to try for contacting a candidate device later
func (d *IntroducedDeviceDetails) CollectAddresses(addresses []string) {
	if len(addresses) == 0 {
		return
	}
	// Sort addresses into a map for deduplication
	addressMap := make(map[string]struct{}, len(d.Addresses))
	for _, s := range d.Addresses {
		addressMap[s] = struct{}{}
	}
	for _, s := range addresses {
		addressMap[s] = struct{}{}
	}
	d.Addresses = make([]string, 0, len(addressMap))
	for a, _ := range addressMap {
		d.Addresses = append(d.Addresses, a)
	}
}
