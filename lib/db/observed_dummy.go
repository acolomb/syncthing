// Only for development testing, will be removed.

package db

import (
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
)

func (db *Lowlevel) CandidateLinksDummy() ([]CandidateLink, error) {
	res := []CandidateLink{
		{
			Introducer: protocol.TestDeviceID2,
			Folder:     "boggl-goggl",
			Candidate:  protocol.TestDeviceID1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "frob"}},
		{
			Introducer: protocol.TestDeviceID2,
			Folder:     "sleep-wells",
			Candidate:  protocol.TestDeviceID1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "nic"}},
		{
			Introducer: protocol.TestDeviceID2,
			Folder:     "damtn-omola",
			Candidate:  protocol.TestDeviceID1,
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

func (db *Lowlevel) CandidateFoldersDummy(device protocol.DeviceID) (map[string]CandidateFolder, error) {
	res := make(map[string]CandidateFolder)
	res["frob"] = CandidateFolder{
		protocol.TestDeviceID1: candidateFolderDevice{
			IntroducedBy: map[protocol.DeviceID]candidateFolderAttribution{
				protocol.TestDeviceID2: candidateFolderAttribution{
					Time:  time.Now(),
					Label: "FROBBY",
				},
			},
		},
	}
	res["dodo"] = CandidateFolder{
		protocol.TestDeviceID2: candidateFolderDevice{
			IntroducedBy: map[protocol.DeviceID]candidateFolderAttribution{
				protocol.TestDeviceID1: candidateFolderAttribution{
					Time:  time.Now(),
					Label: "DODODODO",
				},
			},
		},
	}
	return res, nil
}
