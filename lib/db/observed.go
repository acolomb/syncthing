// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"time"

	"github.com/syncthing/syncthing/lib/db/backend"
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
		Time:            time.Now().Round(time.Second),
		IntroducerLabel: label,
		CandidateMeta:   meta,
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

func (db *Lowlevel) CandidateLinksDummy() ([]CandidateLink, error) {
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

func (db *Lowlevel) CandidateLinksDummyData() {
	dev1, _ := protocol.DeviceIDFromString("P56IOI7-MZJNU2Y-IQGDREY-DM2MGTI-MGL3BXN-PQ6W5BM-TBBZ4TJ-XZWICQ2")
	dev2, _ := protocol.DeviceIDFromString("DOVII4U-SQEEESM-VZ2CVTC-CJM4YN5-QNV7DCU-5U3ASRL-YVFG6TH-W5DV5AA")
	dev3, _ := protocol.DeviceIDFromString("YZJBJFX-RDBL7WY-6ZGKJ2D-4MJB4E7-ZATSDUY-LD6Y3L3-MLFUYWE-AEMXJAC")
	dev4, _ := protocol.DeviceIDFromString("UYGDMA4-TPHOFO5-2VQYDCC-7CWX7XW-INZINQT-LE4B42N-4JUZTSM-IWCSXA4")
	dev5, _ := protocol.DeviceIDFromString("AIBAEAQ-CAIBAEC-AQCAIBA-EAQCAIA-BAEAQCA-IBAEAQC-CAIBAEA-QCAIBA7")

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pics from Jane", dev1, dev2,
		&IntroducedDeviceDetails{
			CertName: "",
			Addresses: []string{
				"192.168.1.2:22000",
				"[2a02:8070:::ff34:1234::aabb]:22000",
			},
			SuggestedName: "Jane",
		}))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pics of J & J", dev1, dev3,
		&IntroducedDeviceDetails{
			CertName: "",
			Addresses: []string{
				"192.168.1.2:22000",
				"[2a02:8070:::ff34:1234::aabb]:22000",
			},
			SuggestedName: "Jane's Laptop",
		}))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Family pics", dev4, dev3, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("abcde-fghij", "Mighty nice folder", dev4, dev3, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Family pics", dev4, dev5, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pictures from Joe", dev2, dev3, nil))
}

func (db *Lowlevel) CandidateLinks() ([]CandidateLink, error) {
	//FIXME not implemented
	return nil, nil
}

func (db *Lowlevel) readCandidateLink(iter backend.Iterator) (ocl ObservedCandidateLink, candidateID, introducerID protocol.DeviceID, folderID string, err error) {
	var deleteCause string
	var bs []byte
	keyDev, ok := db.keyer.IntroducerFromCandidateLinkKey(iter.Key())
	introducerID, err = protocol.DeviceIDFromBytes(keyDev)
	if !ok || err != nil {
		deleteCause = "invalid introducer device ID"
		goto deleteKey
	}
	if keyFolder, ok := db.keyer.FolderFromCandidateLinkKey(iter.Key()); !ok || len(keyFolder) < 1 {
		deleteCause = "invalid folder ID"
		goto deleteKey
	} else {
		folderID = string(keyFolder)
	}
	keyDev = db.keyer.DeviceFromCandidateLinkKey(iter.Key())
	candidateID, err = protocol.DeviceIDFromBytes(keyDev)
	if err != nil {
		deleteCause = "invalid candidate device ID"
		goto deleteKey
	}
	if bs, err = db.Get(iter.Key()); err != nil {
		deleteCause = "DB Get failed"
		goto deleteKey
	}
	if err = ocl.Unmarshal(bs); err != nil {
		deleteCause = "DB Unmarshal failed"
		goto deleteKey
	}
	return

deleteKey:
	l.Infof("Invalid candidate link entry (%v / %v), deleting from database: %x",
		deleteCause, err, iter.Key())
	err = db.Delete(iter.Key())
	return
}

// Consolidated information about a candidate device, enough to add a connection to it
type CandidateDevice struct {
	CertName     string                                           `json:"certName,omitempty"`
	Addresses    []string                                         `json:"addresses,omitempty"`
	IntroducedBy map[protocol.DeviceID]candidateDeviceAttribution `json:"introducedBy"`
}

// Details which an introducer told us about a candidate device
type candidateDeviceAttribution struct {
	Time          time.Time         `json:"time"`
	CommonFolders map[string]string `json:"commonFolders"`
	SuggestedName string            `json:"suggestedName,omitempty"`
}

func (db *Lowlevel) CandidateDevicesDummy(folder string) (map[protocol.DeviceID]CandidateDevice, error) {
	res := make(map[protocol.DeviceID]CandidateDevice)
	res[protocol.TestDeviceID1] = CandidateDevice{
		IntroducedBy: map[protocol.DeviceID]candidateDeviceAttribution{
			protocol.TestDeviceID2: candidateDeviceAttribution{
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
		IntroducedBy: map[protocol.DeviceID]candidateDeviceAttribution{
			protocol.TestDeviceID1: candidateDeviceAttribution{
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

func (db *Lowlevel) CandidateDevices(folder string) (map[protocol.DeviceID]CandidateDevice, error) {
	//db.CandidateLinksDummyData()

	iter, err := db.NewPrefixIterator([]byte{KeyTypeCandidateLink})
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	res := make(map[protocol.DeviceID]CandidateDevice)
	for iter.Next() {
		ocl, candidateID, introducerID, folderID, err := db.readCandidateLink(iter)
		if err != nil {
			return nil, err
		}
		cd, ok := res[candidateID]
		if !ok {
			cd = CandidateDevice{
				Addresses:    []string{},
				IntroducedBy: map[protocol.DeviceID]candidateDeviceAttribution{},
			}
		}
		cd.mergeCandidateLink(ocl, folderID, introducerID)
		res[candidateID] = cd
	}
	return res, nil
}

func (cd *CandidateDevice) mergeCandidateLink(observed ObservedCandidateLink, folder string, introducer protocol.DeviceID) {
	attrib, ok := cd.IntroducedBy[introducer]
	if !ok {
		attrib = candidateDeviceAttribution{
			CommonFolders: map[string]string{},
		}
	}
	attrib.Time = observed.Time
	attrib.CommonFolders[folder] = observed.IntroducerLabel
	if observed.CandidateMeta != nil {
		if cd.CertName != observed.CandidateMeta.CertName {
			//FIXME warn?
			cd.CertName = observed.CandidateMeta.CertName
		}
		cd.collectAddresses(observed.CandidateMeta.Addresses)
		// Only if the device ID is not known locally:
		attrib.SuggestedName = observed.CandidateMeta.SuggestedName
	}
	cd.IntroducedBy[introducer] = attrib
}

// Collect addresses to try for contacting a candidate device later
func (d *CandidateDevice) collectAddresses(addresses []string) {
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

// Consolidated information about a candidate folder
type CandidateFolder map[protocol.DeviceID]map[protocol.DeviceID]candidateFolderAttribution

// Details which an introducer told us about a candidate device
type candidateFolderAttribution struct {
	Time  time.Time `json:"time"`
	Label string    `json:"label"`
}

func (db *Lowlevel) CandidateFoldersDummy() (map[string]CandidateFolder, error) {
	res := make(map[string]CandidateFolder)
	res["frob"] = CandidateFolder{
		protocol.TestDeviceID1: map[protocol.DeviceID]candidateFolderAttribution{
			protocol.TestDeviceID2: candidateFolderAttribution{
				Time:  time.Now(),
				Label: "FROBBY",
			},
		},
	}
	res["dodo"] = CandidateFolder{
		protocol.TestDeviceID2: map[protocol.DeviceID]candidateFolderAttribution{
			protocol.TestDeviceID1: candidateFolderAttribution{
				Time:  time.Now(),
				Label: "DODODODO",
			},
		},
	}
	return res, nil
}

func (db *Lowlevel) CandidateFolders() (map[string]CandidateFolder, error) {
	iter, err := db.NewPrefixIterator([]byte{KeyTypeCandidateLink})
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	res := make(map[string]CandidateFolder)
	for iter.Next() {
		ocl, candidateID, introducerID, folderID, err := db.readCandidateLink(iter)
		if err != nil {
			return nil, err
		}
		cf, ok := res[folderID]
		if !ok {
			cf = make(CandidateFolder)
		}
		cf.mergeCandidateLink(ocl, candidateID, introducerID)
		res[folderID] = cf
		continue
	}
	return res, nil
}

func (cf *CandidateFolder) mergeCandidateLink(observed ObservedCandidateLink, candidate, introducer protocol.DeviceID) {
	attributions, ok := (*cf)[candidate]
	if !ok {
		attributions = make(map[protocol.DeviceID]candidateFolderAttribution)
	}
	attributions[introducer] = candidateFolderAttribution{
		Time:  observed.Time,
		Label: observed.IntroducerLabel,
	}
	(*cf)[candidate] = attributions
}
