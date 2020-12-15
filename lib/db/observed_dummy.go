// Only for development testing, will be removed.

package db

import (
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
)

var (
	testDev1, _ = protocol.DeviceIDFromString("AEAQCAI-BAEAQCA-AIBAEAQ-CAIBAEC-AQCAIBA-EAQCAIA-BAEAQCA-IBAEAQC")
	testDev2, _ = protocol.DeviceIDFromString("AIBAEAQ-CAIBAEC-AQCAIBA-EAQCAIA-BAEAQCA-IBAEAQC-CAIBAEA-QCAIBA7")
	testDev3, _ = protocol.DeviceIDFromString("P56IOI7-MZJNU2Y-IQGDREY-DM2MGTI-MGL3BXN-PQ6W5BM-TBBZ4TJ-XZWICQ2")
	testDev4, _ = protocol.DeviceIDFromString("DOVII4U-SQEEESM-VZ2CVTC-CJM4YN5-QNV7DCU-5U3ASRL-YVFG6TH-W5DV5AA")
	testDev5, _ = protocol.DeviceIDFromString("YZJBJFX-RDBL7WY-6ZGKJ2D-4MJB4E7-ZATSDUY-LD6Y3L3-MLFUYWE-AEMXJAC")
	testDev6, _ = protocol.DeviceIDFromString("UYGDMA4-TPHOFO5-2VQYDCC-7CWX7XW-INZINQT-LE4B42N-4JUZTSM-IWCSXA4")
)

func (db *Lowlevel) CandidateLinksDummy() ([]CandidateLink, error) {
	res := []CandidateLink{
		{
			Introducer: testDev2,
			Folder:     "boggl-goggl",
			Candidate:  testDev1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "frob"}},
		{
			Introducer: testDev2,
			Folder:     "sleep-wells",
			Candidate:  testDev1,
			ObservedCandidateLink: ObservedCandidateLink{
				Time:            time.Now().Round(time.Second),
				IntroducerLabel: "nic"}},
		{
			Introducer: testDev2,
			Folder:     "damtn-omola",
			Candidate:  testDev1,
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
	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pics from Jane", testDev3, testDev4,
		&IntroducedDeviceDetails{
			CertName: "",
			Addresses: []string{
				"192.168.1.2:22000",
				"[2a02:8070:::ff34:1234::aabb]:22000",
			},
			SuggestedName: "Jane",
		}))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pics of J & J", testDev3, testDev5,
		&IntroducedDeviceDetails{
			CertName: "",
			Addresses: []string{
				"192.168.1.2:22000",
				"[2a02:8070:::ff34:1234::aabb]:22000",
			},
			SuggestedName: "Jane's Laptop",
		}))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Family pics", testDev6, testDev5, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("abcde-fghij", "Mighty nice folder", testDev6, testDev5, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Family pics", testDev6, testDev2, nil))

	l.Warnln(db.AddOrUpdateCandidateLink("cpkn4-57ysy", "Pictures from Joe", testDev4, testDev5, nil))
}

func (db *Lowlevel) CandidateDevicesDummy() (map[protocol.DeviceID]CandidateDevice, error) {
	res := map[protocol.DeviceID]CandidateDevice{
		testDev1: CandidateDevice{
			IntroducedBy: map[protocol.DeviceID]candidateDeviceAttribution{
				testDev2: candidateDeviceAttribution{
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
		},
		testDev2: CandidateDevice{
			IntroducedBy: map[protocol.DeviceID]candidateDeviceAttribution{
				testDev1: candidateDeviceAttribution{
					Time: time.Now(),
					CommonFolders: map[string]string{
						"dodo": "DODODODO",
					},
					SuggestedName: "coolnhip",
				},
			},
			CertName:  "syncthing",
			Addresses: []string{"bar", "baz"},
		},
	}
	return res, nil
}

func (db *Lowlevel) CandidateFoldersDummy() (map[string]CandidateFolder, error) {
	res := map[string]CandidateFolder{
		"frob": CandidateFolder{
			testDev1: candidateFolderDevice{
				IntroducedBy: map[protocol.DeviceID]candidateFolderAttribution{
					testDev2: candidateFolderAttribution{
						Time:  time.Now(),
						Label: "FROBBY",
					},
				},
			},
		},
		"dodo": CandidateFolder{
			testDev2: candidateFolderDevice{
				IntroducedBy: map[protocol.DeviceID]candidateFolderAttribution{
					testDev1: candidateFolderAttribution{
						Time:  time.Now(),
						Label: "DODODODO",
					},
				},
			},
		},
	}
	return res, nil
}
