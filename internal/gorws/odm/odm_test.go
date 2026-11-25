package odm

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStudyList(t *testing.T) {
	str := `<?xml version="1.0" encoding="utf-8"?>
<ODM FileType="Snapshot" FileOID="0bc7a538-c09d-4a1a-970d-ee9c6ee07cd0" CreationDateTime="2024-04-02T08:31:40.362-00:00" ODMVersion="1.3" xmlns:mdsol="http://www.mdsol.com/ns/odm/metadata" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.cdisc.org/ns/odm/v1.3">
    <Study OID="CA-4948-104(Prod)">
        <GlobalVariables>
            <StudyName>CA-4948-104</StudyName>
            <StudyDescription />
            <ProtocolName>CA-4948-104</ProtocolName>
        </GlobalVariables>
    </Study>
    <Study OID="CA-4948-104(DEV)">
        <GlobalVariables>
            <StudyName>CA-4948-104 (DEV)</StudyName>
            <StudyDescription />
            <ProtocolName>CA-4948-104</ProtocolName>
        </GlobalVariables>
    </Study>
</ODM>`
	odm := ODM{}
	err := xml.Unmarshal([]byte(str), &odm)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, ODM{
		Studies: []*Study{
			{
				GlobalVariables: &GlobalVariables{
					StudyName:        "CA-4948-104",
					StudyDescription: "",
					ProtocolName:     "CA-4948-104",
				},
				OID: "CA-4948-104(Prod)",
			},
			{
				GlobalVariables: &GlobalVariables{
					StudyName:        "CA-4948-104 (DEV)",
					StudyDescription: "",
					ProtocolName:     "CA-4948-104",
				},
				OID: "CA-4948-104(DEV)",
			},
		},
		FileType:         "Snapshot",
		FileOID:          "0bc7a538-c09d-4a1a-970d-ee9c6ee07cd0",
		CreationDateTime: "2024-04-02T08:31:40.362-00:00",
		ODMVersion:       "1.3",
		XMLNs:            "http://www.cdisc.org/ns/odm/v1.3",
	}, odm, "they should be equal")
}

func TestParseClinicalDataList(t *testing.T) {
	str := `<?xml version="1.0" encoding="utf-8"?>
<ODM FileType="Snapshot" FileOID="c74671d4-0fe0-4534-9912-617648262a8e" CreationDateTime="2024-04-07T02:02:37.464-00:00" ODMVersion="1.3" xmlns:mdsol="http://www.mdsol.com/ns/odm/metadata" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.cdisc.org/ns/odm/v1.3">
    <ClinicalData StudyOID="CA-4948-104(DEV)" MetaDataVersionOID="4">
        <SubjectData SubjectKey="1001-0001" mdsol:SubjectKeyType="SubjectName">
            <SiteRef LocationOID="DEV-1001" mdsol:StudyEnvSiteNumber="1001" />
        </SubjectData>
    </ClinicalData>
    <ClinicalData StudyOID="CA-4948-104(DEV)" MetaDataVersionOID="4">
        <SubjectData SubjectKey="1001-0002" mdsol:SubjectKeyType="SubjectName">
            <SiteRef LocationOID="DEV-1001" mdsol:StudyEnvSiteNumber="1001" />
        </SubjectData>
    </ClinicalData>
</ODM>`
	odm := ODM{}
	err := xml.Unmarshal([]byte(str), &odm)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, ODM{
		ClinicalData: []*ClinicalData{
			{
				StudyOID:           "CA-4948-104(DEV)",
				MetaDataVersionOID: "4",
				SubjectData: []*SubjectData{
					{
						SubjectKey:     "1001-0001",
						SubjectKeyType: "SubjectName",
						SiteRef: &SiteRef{
							LocationOID:        "DEV-1001",
							StudyEnvSiteNumber: "1001",
						},
					},
				},
			},
			{
				StudyOID:           "CA-4948-104(DEV)",
				MetaDataVersionOID: "4",
				SubjectData: []*SubjectData{
					{
						SubjectKey:     "1001-0002",
						SubjectKeyType: "SubjectName",
						SiteRef: &SiteRef{
							LocationOID:        "DEV-1001",
							StudyEnvSiteNumber: "1001",
						},
					},
				},
			},
		},
		FileType:         "Snapshot",
		FileOID:          "c74671d4-0fe0-4534-9912-617648262a8e",
		CreationDateTime: "2024-04-07T02:02:37.464-00:00",
		ODMVersion:       "1.3",
		XMLNs:            "http://www.cdisc.org/ns/odm/v1.3",
	}, odm, "they should be equal")
}
